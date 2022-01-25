package gotwilio

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

// These are the parameters to use when you are requesting account usage.
// See https://www.twilio.com/docs/usage/api/usage-record#read-multiple-usagerecord-resources
// for more info.
type UsageParameters struct {
	Category           string // Optional
	StartDate          string // Optional, in YYYY-MM-DD or as offset
	EndDate            string // Optional, in YYYY-MM-DD or as offset
	IncludeSubaccounts bool   // Optional
}

// UsageRecord specifies the usage for a particular usage category.
// See https://www.twilio.com/docs/usage/api/usage-record#usagerecord-properties
// for more info.
type UsageRecord struct {
	AccountSid  string `json:"account_sid"`
	Category    string `json:"category"`
	Description string `json:"description"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Price       string `json:"price"`
	PriceUnit   string `json:"price_unit"`
	Count       int    `json:"count,string"`
	CountUnit   string `json:"count_unit"`
	Usage       string `json:"usage"`
	UsageUnit   string `json:"usage_unit"`
	AsOf        string `json:"as_of"` // GMT timestamp formatted as YYYY-MM-DDTHH:MM:SS+00:00
	// TODO: handle SubresourceUris
}

// UsageResponse contains information about account usage.
type UsageResponse struct {
	PageSize     int           `json:"page_size"`
	Page         int           `json:"page"`
	UsageRecords []UsageRecord `json:"usage_records"`
	NextPageUri  string        `json:"next_page_uri"`
}

func (twilio *Twilio) GetUsage(category, startDate, endDate string, includeSubaccounts bool) ([]UsageRecord, *Exception, error) {
	return twilio.GetUsageWithContext(context.Background(), category, startDate, endDate, includeSubaccounts)
}

func (twilio *Twilio) GetUsageWithContext(ctx context.Context, category, startDate, endDate string, includeSubaccounts bool) ([]UsageRecord, *Exception, error) {
	formValues := url.Values{}
	if category != "" {
		formValues.Set("Category", category)
	}
	if startDate != "" {
		formValues.Set("StartDate", startDate)
	}
	if endDate != "" {
		formValues.Set("EndDate", endDate)
	}
	formValues.Set("IncludeSubaccounts", strconv.FormatBool(includeSubaccounts))

	var usageResponse *UsageResponse
	var exception *Exception
	var usageRecords []UsageRecord

	for {
		if usageResponse != nil && usageResponse.NextPageUri == "" {
			break
		}

		twilioUrl := twilio.BaseUrl + "/Accounts/" + twilio.AccountSid + "/Usage/Records.json?" + formValues.Encode()
		if usageResponse != nil && usageResponse.NextPageUri != "" {
			// clean up "/2010-04-01" that appears at the end of twilio.BaseUrl and beginning of each NextPageUri
			uri := strings.Replace(usageResponse.NextPageUri, path.Base(twilio.BaseUrl), "", 1)
			twilioUrl = twilio.BaseUrl + path.Clean(uri)
		}

		res, err := twilio.get(ctx, twilioUrl)
		if err != nil {
			return nil, nil, err
		}
		defer res.Body.Close()

		usageResponse, exception, err = parseResponse(res)
		if exception != nil || err != nil {
			return nil, exception, err
		}
		usageRecords = append(usageRecords, usageResponse.UsageRecords...)
	}

	return usageRecords, nil, nil
}

func parseResponse(res *http.Response) (*UsageResponse, *Exception, error) {
	responseBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, nil, err
	}
	if res.StatusCode != http.StatusOK {
		exception := new(Exception)
		err = json.Unmarshal(responseBody, exception)
		return nil, exception, err
	}

	usageResponse := new(UsageResponse)
	err = json.Unmarshal(responseBody, usageResponse)
	return usageResponse, nil, err
}
