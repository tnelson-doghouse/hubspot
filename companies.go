package hubspot

import (
	"fmt"
	"strconv"
	"regexp"
)

// Companies client
type Companies struct {
	Client
	objectPath string
}

// Companies constructor (from Client)
func (c Client) Companies() Companies {
	return Companies{
		Client: c,
		objectPath: c.objectPath("companies", "", "v3"),
	}
}

// CompaniesRequest object
type CompaniesRequest struct {
	Properties []Property `json:"properties"`
}

// CompaniesResponse object
type CompaniesResponse struct {
	ErrorResponse
	CompanyID  int               `json:"companyId"`
	PortalID   int               `json:"portalId"`
	Properties CompanyProperties `json:"properties"`
	IsDeleted  bool              `json:"isDeleted"`
}

type PagingResponsePage struct {
	After  string  `json:after`
	Link   string  `json:link`
}

type PagingResponse struct {
	Next   PagingResponsePage   `json:next`
}

type CompaniesListResponse struct {
	Results       []CompaniesResponse  `json:results`
	Paging        PagingResponse   `json:paging`
}

// CompanyProperties response object
type CompanyProperties struct {
	CreateDate  string `json:"createdate"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// List Companies constructor
// If relative = true, then nextlink = a relative path that needs to be run through objectPath
// If relative = false, then nextlink = a full path (protocol and hostname will be stripped away)
func (c Client) CompaniesList(nextlink string, relative bool) (CompaniesListResponse, error) {
	path := ""
	if relative {
		path = c.objectPath("companies", path, "v3")
	} else {
		pattern := regexp.MustCompile("^https?://[^/]+")
		path = pattern.ReplaceAllString(nextlink, "")
	}
	r := CompaniesListResponse{}
	return r, fmt.Errorf("hubspot.Companies.CompaniesList(): Using path %s", path)
	err := c.Request("GET", path, nil, &r)
	if err != nil {
		fmt.Print(err)
	}
	return r, err
}


// Get Company
func (c Companies) Get(companyID int) (CompaniesResponse, error) {
	r := CompaniesResponse{}
	err := c.Client.Request("GET", c.objectPath + "/"+strconv.Itoa(companyID), nil, &r)
	return r, err
}

// Create new Company
func (c Companies) Create(data CompaniesRequest) (CompaniesResponse, error) {
	r := CompaniesResponse{}
	err := c.Client.Request("POST", c.objectPath + "/", data, &r)
	return r, err
}

// Update Deal
func (c Companies) Update(companyID int, data CompaniesRequest) (CompaniesResponse, error) {
	r := CompaniesResponse{}
	err := c.Client.Request("PUT", c.objectPath + "/"+strconv.Itoa(companyID), data, &r)
	return r, err
}

// Delete Deal
func (c Companies) Delete(companyID int) error {
	err := c.Client.Request("DELETE", c.objectPath + "/"+strconv.Itoa(companyID), nil, nil)
	return err
}

