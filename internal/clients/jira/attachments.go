package jira

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"buddy/internal/errors"
)

// GetAttachmentContent downloads attachment content
func (c *JiraClient) GetAttachmentContent(ctx context.Context, attachmentURL string) ([]byte, error) {
	client := &http.Client{
		Timeout: c.httpClient.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			c.setAuthHeaders(req)
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", attachmentURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeExternal, "failed to create attachment request")
	}

	c.setAuthHeaders(req)

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeExternal, "failed to download attachment")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleAPIError(resp)
	}

	content, err := readResponseBody(resp)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeExternal, "failed to read attachment content")
	}

	return content, nil
}

// DownloadAttachment downloads an attachment and saves it to the specified path
func (c *JiraClient) DownloadAttachment(ctx context.Context, attachment Attachment, savePath string) error {
	content, err := c.GetAttachmentContent(ctx, attachment.URL)
	if err != nil {
		return errors.Wrap(err, errors.ErrorTypeExternal, "failed to download attachment")
	}

	finalPath := c.resolveSavePath(savePath, attachment.Filename)

	if err := os.WriteFile(finalPath, content, 0644); err != nil {
		return errors.Wrap(err, errors.ErrorTypeInternal, "failed to save attachment")
	}

	c.logger.Info("Downloaded attachment: %s", finalPath)
	return nil
}

// resolveSavePath resolves the final save path, handling directories and duplicate files
func (c *JiraClient) resolveSavePath(savePath, filename string) string {
	// Handle file naming if savePath is a directory
	if strings.HasSuffix(savePath, "/") || savePath == "." || savePath == "" {
		fullPath := filepath.Join(savePath, filename)

		// If file already exists, add suffix
		if _, err := os.Stat(fullPath); err == nil {
			ext := filepath.Ext(filename)
			base := strings.TrimSuffix(filename, ext)
			i := 1
			for {
				newFilename := fmt.Sprintf("%s_%d%s", base, i, ext)
				newPath := filepath.Join(savePath, newFilename)
				if _, err := os.Stat(newPath); os.IsNotExist(err) {
					return newPath
				}
				i++
			}
		}
		return fullPath
	}

	return savePath
}
