package gotwilio

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type IceServer struct {
	Credential string `json:"credential"`
	Url        string `json:"url"`
	Urls       string `json:"urls"`
	Username   string `json:"username"`
}

type NtsResponse struct {
	AccountSid  string
	DateCreated time.Time
	DateUpdated time.Time
	IceServers  []*IceServer
	Password    string
	Ttl         int
	Username    string
}

func (nts *NtsResponse) UnmarshalJSON(b []byte) error {
	var raw map[string]json.RawMessage
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}
	for k, v := range raw {
		switch k {
		case "date_created":
			nts.DateCreated, err = time.Parse(`"`+time.RFC1123Z+`"`, string(v))
			if err != nil {
				return err
			}
		case "date_updated":
			nts.DateUpdated, err = time.Parse(`"`+time.RFC1123Z+`"`, string(v))
			if err != nil {
				return err
			}
		case "ice_servers":
			err = json.Unmarshal(v, &nts.IceServers)
			if err != nil {
				return err
			}
		case "ttl":
			nts.Ttl, err = strconv.Atoi(strings.Trim(string(v), "\""))
			if err != nil {
				return err
			}
		case "account_sid":
			nts.AccountSid = strings.Trim(string(v), "\"")
		case "username":
			nts.Username = strings.Trim(string(v), "\"")
		case "password":
			nts.Password = strings.Trim(string(v), "\"")
		}
	}
	return nil
}

func (twilio *Twilio) GetNtsToken(ttl int) (ntsResponse *NtsResponse, exception *Exception, err error) {
	basePath := "https://api.twilio.com/2010-04-01"
	twilioUrl := fmt.Sprintf("%s/Accounts/%s/Tokens.json?Ttl=%d", basePath, twilio.AccountSid, ttl)

	res, err := twilio.post(nil, twilioUrl)
	if err != nil {
		return ntsResponse, exception, err
	}
	defer res.Body.Close()

	responseBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ntsResponse, exception, err
	}

	if res.StatusCode != http.StatusCreated {
		exception = new(Exception)
		err = json.Unmarshal(responseBody, exception)

		// We aren't checking the error because we don't actually care.
		// It's going to be passed to the client either way.
		return ntsResponse, exception, err
	}

	ntsResponse = new(NtsResponse)
	err = json.Unmarshal(responseBody, ntsResponse)
	if err != nil {
		log.Println(err)
		log.Println(string(responseBody))
	}

	return ntsResponse, exception, err
}
