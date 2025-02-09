package fetcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"golang.org/x/sync/errgroup"

	"github.com/ttl256/euivator/internal/registry"
)

const ETagsFile = "etags.json"

type RespURLHeader struct {
	URL    string
	Header http.Header
}

type ETagStorage map[string]string

type Fetcher struct {
	Sources   []Source
	Dir       string
	ETagsFile string
	Logger    *slog.Logger
}

func New(sources []Source, dir string, logger *slog.Logger) *Fetcher {
	return &Fetcher{
		Sources:   sources,
		Dir:       dir,
		ETagsFile: filepath.Join(dir, ETagsFile),
		Logger:    logger,
	}
}

func (s *Fetcher) SaveETags(tags ETagStorage) error {
	data, err := json.MarshalIndent(tags, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling data: %w", err)
	}

	f, err := os.OpenFile(
		s.ETagsFile,
		os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return fmt.Errorf("opening a file: %w", err)
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("writing to a file: %w", err)
	}

	return nil
}

func (s *Fetcher) LoadETags() (ETagStorage, error) {
	data, err := os.ReadFile(s.ETagsFile)
	if err != nil {
		return nil, fmt.Errorf("reading etags file: %w", err)
	}

	tags := new(ETagStorage)
	err = json.Unmarshal(data, tags)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling etags: %w", err)
	}

	return *tags, nil
}

/*
DownloadFiles fetches OUI registry files. The download is omitted when all of
the following conditions are true: 1) a file already exists locally 2) the use
of ETags is enabled 3) ETags exist.
*/
//nolint: gocognit // fine
func (s *Fetcher) DownloadFiles(ctx context.Context, useETags bool) error {
	etags, err := s.LoadETags()
	if err != nil {
		// Ignore absence of etags file.
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	g, ctxGroup := errgroup.WithContext(ctx)

	if err = os.MkdirAll(s.Dir, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	respHeaderCh := make(chan RespURLHeader)

	for _, source := range s.Sources {
		g.Go(func() error {
			filepath := filepath.Join(s.Dir, strings.Join([]string{string(source.RegistryName), "csv"}, "."))

			var fileExists = true
			_, err = os.Stat(filepath)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					fileExists = false
				} else {
					return fmt.Errorf("checking whether %q exists: %w", filepath, err)
				}
			}

			var f *os.File
			f, err = os.OpenFile(
				filepath,
				os.O_CREATE|os.O_WRONLY,
				0644,
			)
			if err != nil {
				return fmt.Errorf("opening a file: %w", err)
			}
			defer f.Close()

			header := make(http.Header)
			if fileExists && useETags {
				if etag, ok := etags[source.URL.String()]; ok {
					header.Set("If-None-Match", etag)
				}
			}
			var respHeader http.Header
			respHeader, err = FetchFile(ctxGroup, source.URL, &header, f, s.Logger)
			if err != nil {
				return err
			}
			s.Logger.LogAttrs(ctx, slog.LevelDebug, "file saved", slog.String("path", filepath))

			select {
			case <-ctxGroup.Done():
				return ctxGroup.Err()
			case respHeaderCh <- RespURLHeader{URL: source.URL.String(), Header: respHeader}:
			}

			return nil
		})
	}

	go func() {
		_ = g.Wait()
		close(respHeaderCh)
	}()

	tags := make(ETagStorage)
	for header := range respHeaderCh {
		if etag := header.Header.Get("ETag"); etag != "" {
			tags[header.URL] = etag
		}
	}

	err = g.Wait()
	if err != nil {
		return fmt.Errorf("getting files: %w", err)
	}

	err = s.SaveETags(tags)
	if err != nil {
		return fmt.Errorf("saving etags: %w", err)
	}

	return nil
}

type Source struct {
	URL          url.URL
	RegistryName registry.Name
}

func GetSources() []Source {
	uncheckedSources := []struct {
		URL          string
		RegistryName registry.Name
	}{
		{URL: "https://standards-oui.ieee.org/oui/oui.csv", RegistryName: registry.NameMAL},
		{URL: "https://standards-oui.ieee.org/oui28/mam.csv", RegistryName: registry.NameMAM},
		{URL: "https://standards-oui.ieee.org/oui36/oui36.csv", RegistryName: registry.NameMAS},
		{URL: "https://standards-oui.ieee.org/cid/cid.csv", RegistryName: registry.NameCID},
	}

	sources := make([]Source, 0, len(uncheckedSources))
	for _, source := range uncheckedSources {
		URL, err := url.Parse(source.URL) //nolint: gocritic // avoid collisions with url.URL
		if err != nil {
			panic(fmt.Errorf("unable to parse URL %q: %w", URL, err))
		}

		sources = append(sources, Source{URL: *URL, RegistryName: source.RegistryName})
	}

	return sources
}

/**/
//nolint: gocritic // avoid collisions with url.URL
func FetchFile(
	ctx context.Context, URL url.URL, headers *http.Header, w io.Writer, logger *slog.Logger,
) (http.Header, error) {
	logger.LogAttrs(ctx, slog.LevelInfo, "fetching", slog.String("url", URL.String()))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, URL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if headers != nil {
		for k, v := range *headers {
			for _, vv := range v {
				req.Header.Add(k, vv)
			}
		}
	}
	logger.LogAttrs(
		ctx,
		slog.LevelDebug,
		"starting download",
		slog.String("url", URL.String()),
		slog.Any("headers", req.Header),
	)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting response: %w", err)
	}
	defer resp.Body.Close()

	logger.LogAttrs(
		ctx,
		slog.LevelDebug,
		"response",
		slog.String("url", URL.String()),
		slog.Int("code", resp.StatusCode),
		slog.Any("header", resp.Header),
	)

	switch resp.StatusCode {
	case http.StatusOK:
		var n int64

		n, err = io.Copy(w, resp.Body)
		if err != nil {
			return nil, fmt.Errorf("writing response: %w", err)
		}

		logger.LogAttrs(
			ctx,
			slog.LevelInfo,
			"finished download",
			slog.String("url", URL.String()),
			slog.String("size", humanize.Bytes(uint64(n))), //nolint: gosec // n is always non-negative
		)
	case http.StatusNotModified:
		logger.LogAttrs(
			ctx,
			slog.LevelInfo,
			"omitting download",
			slog.String("url", URL.String()),
		)
	default:
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.Header, nil
}
