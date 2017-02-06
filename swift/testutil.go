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

type MockedTestSet struct {
	Account       *Account
	Connection    *Connection
	Container     *LogicalContainer
	ContainerList ContainerList
}

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

func NewMockedTestSet() *MockedTestSet {
	account := &Account{
		&lib.Account{
			Containers: 2,
			Objects:    1500,
			BytesUsed:  65536,
		},
		lib.Headers{
			AccountBytesUsedHeader:      "65536",
			AccountContainerCountHeader: "2",
			AccountObjectCountHeader:    "1500",
		},
	}
	connection := &Connection{
		Connection: &lib.Connection{
			AuthToken:  MockedToken,
			StorageUrl: MockedStorageURL,
			Transport:  httpmock.DefaultTransport,
		},
		StoragePolicy: "Policy1",
	}
	container := &LogicalContainer{
		MainContainer: &Container{
			Container: &lib.Container{
				Name:  "container",
				Bytes: 16384,
				Count: 200,
			},
			Headers: lib.Headers{
				StoragePolicyHeader:        "Policy1",
				ContainerBytesUsedHeader:   "16384",
				ContainerObjectCountHeader: "200",
			},
		},
		SegmentContainer: &Container{
			Container: &lib.Container{
				Name:  "container_segments",
				Bytes: 32768,
				Count: 500,
			},
			Headers: lib.Headers{
				StoragePolicyHeader:        "Policy1",
				ContainerBytesUsedHeader:   "32768",
				ContainerObjectCountHeader: "500",
			},
		},
	}
	containerList := ContainerList{
		"container":          container.MainContainer,
		"container_segments": container.SegmentContainer,
	}

	return &MockedTestSet{
		Account:       account,
		Connection:    connection,
		Container:     container,
		ContainerList: containerList,
	}
}

func (ts *MockedTestSet) MockAccount(status StatusMap) {
	// Account information
	httpmock.RegisterResponder("HEAD", MockedStorageURL,
		func(req *http.Request) (resp *http.Response, err error) {
			resp = httpmock.NewStringResponse(status["HEAD"], "")

			if status["HEAD"] < 200 || status["HEAD"] >= 300 {
				return
			}

			addInt64Header(resp, AccountBytesUsedHeader, ts.Account.BytesUsed)
			addInt64Header(resp, AccountContainerCountHeader, ts.Account.Containers)
			addInt64Header(resp, AccountObjectCountHeader, ts.Account.Objects)
			if ts.Account.Quota > 0 {
				addInt64Header(resp, AccountBytesQuotaHeader, ts.Account.Quota)
			}

			return
		},
	)

	// Container list
	httpmock.RegisterResponder("GET", MockedStorageURL,
		func(req *http.Request) (resp *http.Response, err error) {
			body, err := json.Marshal(ts.ContainerList.Slice())
			if err != nil {
				return
			}
			resp = httpmock.NewBytesResponse(status["GET"], body)
			return
		},
	)

	// Container creation
	for name := range ts.ContainerList {
		httpmock.RegisterResponder("PUT", MockedStorageURL+"/"+name,
			func(req *http.Request) (resp *http.Response, err error) {
				resp = httpmock.NewStringResponse(status["PUT"], "")
				return
			},
		)
	}
}

func (ts *MockedTestSet) MockContainers(status StatusMap) {
	for name, container := range ts.ContainerList {
		c := container
		httpmock.RegisterResponder("HEAD", MockedStorageURL+"/"+name,
			func(req *http.Request) (resp *http.Response, err error) {
				code := status["HEAD"]
				if c == nil {
					code = 404
				}
				resp = httpmock.NewStringResponse(code, "")

				if code < 200 || code >= 300 {
					return
				}

				addInt64Header(resp, ContainerBytesUsedHeader, c.Bytes)
				addInt64Header(resp, ContainerObjectCountHeader, c.Count)
				resp.Header.Add(StoragePolicyHeader,
					c.Headers[StoragePolicyHeader])

				return
			},
		)
		httpmock.RegisterResponder("DELETE", MockedStorageURL+"/"+name,
			func(req *http.Request) (resp *http.Response, err error) {
				code := status["DELETE"]
				if c == nil {
					code = 404
				}
				resp = httpmock.NewStringResponse(code, "")

				return
			},
		)
	}
}
