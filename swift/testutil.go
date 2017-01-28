package swift

import (
	"encoding/json"
	"fmt"
	"net/http"

	lib "github.com/xlucas/swift"
	httpmock "gopkg.in/jarcoal/httpmock.v1"
)

const (
	MockedStorageURL = "https://provider/v1/a"
	MockedToken      = "token"
)

type StatusMap map[string]int

func addInt64Header(resp *http.Response, name string, value int64) {
	resp.Header.Add(name, fmt.Sprintf("%d", value))
}

func NewMockedConnectionHolder(capacity uint32, storagePolicy string) (
	mockedConnectionHolder *ResourceHolder,
) {
	return NewResourceHolder(capacity, &Connection{
		Connection: &lib.Connection{
			AuthToken:  MockedToken,
			StorageUrl: MockedStorageURL,
			Transport:  httpmock.DefaultTransport,
		},
		StoragePolicy: storagePolicy,
	})
}

func MockAccount(account *Account, list ContainerList, status StatusMap) {
	// Account information
	httpmock.RegisterResponder("HEAD", MockedStorageURL,
		func(req *http.Request) (resp *http.Response, err error) {
			resp = httpmock.NewStringResponse(status["HEAD"], "")

			if status["HEAD"] < 200 || status["HEAD"] >= 300 {
				return
			}

			addInt64Header(resp, "X-Account-Bytes-Used", account.BytesUsed)
			addInt64Header(resp, "X-Account-Container-Count", account.Containers)
			addInt64Header(resp, "X-Account-Object-Count", account.Objects)
			addInt64Header(resp, "X-Account-Meta-Quota-Bytes", account.Quota)

			return
		},
	)

	// Container list
	httpmock.RegisterResponder("GET", MockedStorageURL,
		func(req *http.Request) (resp *http.Response, err error) {
			body, err := json.Marshal(list.Slice())
			if err != nil {
				return
			}
			resp = httpmock.NewBytesResponse(status["GET"], body)
			return
		},
	)

	// Container creation
	for name, _ := range list {
		httpmock.RegisterResponder("PUT", MockedStorageURL+"/"+name,
			func(req *http.Request) (resp *http.Response, err error) {
				resp = httpmock.NewStringResponse(status["PUT"], "")
				return
			},
		)
	}
}

func MockContainers(list ContainerList, status StatusMap) {
	for name, container := range list {
		c := container
		httpmock.RegisterResponder("HEAD", MockedStorageURL+"/"+name,
			func(req *http.Request) (resp *http.Response, err error) {
				code := status["HEAD"]
				if c == nil {
					code = 404
				}
				resp = httpmock.NewStringResponse(code, "")

				if status["HEAD"] < 200 || status["HEAD"] >= 300 {
					return
				}

				addInt64Header(resp, "X-Container-Bytes-Used", c.Bytes)
				addInt64Header(resp, "X-Container-Object-Count", c.Count)
				resp.Header.Add(StoragePolicyHeader,
					c.Headers[StoragePolicyHeader])

				return
			},
		)
	}
}
