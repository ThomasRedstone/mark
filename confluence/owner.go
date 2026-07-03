package confluence

import (
	"fmt"
	"net/http"
)

// pageV2 is the subset of the Confluence Cloud v2 page representation needed
// to read and rewrite page ownership.
type pageV2 struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Title   string `json:"title"`
	OwnerID string `json:"ownerId"`

	Version struct {
		Number int64 `json:"number"`
	} `json:"version"`

	Body struct {
		Storage struct {
			Representation string `json:"representation"`
			Value          string `json:"value"`
		} `json:"storage"`
	} `json:"body"`
}

// GetPageOwner returns the current owner accountId of a page via the v2 API.
func (api *API) GetPageOwner(pageID string) (string, error) {
	page, err := api.getPageV2(pageID, false)
	if err != nil {
		return "", err
	}
	return page.OwnerID, nil
}

// EnsurePageOwner transfers ownership of a page to the given accountId.
// Returns false without writing when the page is already owned by that
// account. Ownership is only writable through the v2 update-page endpoint,
// which requires the full page body, so the current storage body is read
// back and re-submitted verbatim — the content itself is untouched.
//
// Only Confluence Cloud supports page ownership; on Server/DC the v2
// endpoint does not exist and this returns an error.
func (api *API) EnsurePageOwner(pageID string, ownerID string) (bool, error) {
	page, err := api.getPageV2(pageID, true)
	if err != nil {
		return false, fmt.Errorf("unable to read page via v2 API: %w", err)
	}

	if page.OwnerID == ownerID {
		return false, nil
	}

	payload := map[string]any{
		"id":     page.ID,
		"status": page.Status,
		"title":  page.Title,
		"body": map[string]any{
			"representation": "storage",
			"value":          page.Body.Storage.Value,
		},
		"version": map[string]any{
			"number":  page.Version.Number + 1,
			"message": "mark: transfer page ownership",
		},
		"ownerId": ownerID,
	}

	request, err := api.restV2.
		Res("pages").
		Id(pageID, &map[string]any{}).
		Put(payload)
	if err != nil {
		return false, err
	}

	if request.Raw.StatusCode != http.StatusOK {
		return false, newErrorStatusNotOK(request)
	}

	return true, nil
}

func (api *API) getPageV2(pageID string, withBody bool) (*pageV2, error) {
	query := map[string]string{}
	if withBody {
		query["body-format"] = "storage"
	}

	request, err := api.restV2.
		Res("pages").
		Id(pageID, &pageV2{}).
		Get(query)
	if err != nil {
		return nil, err
	}

	if request.Raw.StatusCode != http.StatusOK {
		return nil, newErrorStatusNotOK(request)
	}

	return request.Response.(*pageV2), nil
}
