package logblobfinder

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/tmeadon/nsgpeek/internal/nsgpeektest"
	"github.com/tmeadon/nsgpeek/pkg/azure"
)

func TestFindSpecific(t *testing.T) {
	fakeNsgName := "nsg-view"
	fakeBlob := new(azure.Blob)

	type test struct {
		start    time.Time
		end      time.Time
		blobs    []string
		expected []string
		err      error
	}

	tests := []test{
		{
			start: time.Date(2022, 01, 01, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2022, 01, 02, 0, 0, 0, 0, time.UTC),
			blobs: []string{
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=01/h=13/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=01/h=19/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=02/h=00/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=02/h=02/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=01/h=02/m=00/macAddress=0022483F762A/PT1H.json", "bad"),
			},
			expected: []string{
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=01/h=13/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=01/h=19/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=02/h=00/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
			},
			err: nil,
		},
		{
			start: time.Date(2022, 01, 01, 15, 0, 0, 0, time.UTC),
			end:   time.Date(2022, 01, 02, 0, 0, 0, 0, time.UTC),
			blobs: []string{
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=01/h=13/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=01/h=19/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=02/h=00/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=02/h=02/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=01/h=02/m=00/macAddress=0022483F762A/PT1H.json", "bad"),
			},
			expected: []string{
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=01/h=19/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=02/h=00/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
			},
			err: nil,
		},
		{
			start: time.Date(2022, 01, 02, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2022, 01, 01, 0, 0, 0, 0, time.UTC),
			blobs: []string{
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=01/h=13/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=01/h=19/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=02/h=00/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=02/h=02/m=00/macAddress=0022483F762A/PT1H.json", fakeNsgName),
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=01/h=02/m=00/macAddress=0022483F762A/PT1H.json", "bad"),
			},
			expected: []string{},
		},
		{
			start: time.Date(2022, 01, 01, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2022, 01, 02, 0, 0, 0, 0, time.UTC),
			blobs: []string{
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=02/h=00/m=00/macAddress=0022483F762A/PT1H.json", "bad"),
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=02/h=02/m=00/macAddress=0022483F762A/PT1H.json", "bad"),
				fmt.Sprintf("/resourceId=/SUBSCRIPTIONS/xyz/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/%v/y=2022/m=01/d=01/h=02/m=00/macAddress=0022483F762A/PT1H.json", "bad"),
			},
			expected: []string{},
			err:      ErrBlobPrefixNotFound,
		},
	}

	for _, test := range tests {
		mockStorageBlobGetter := nsgpeektest.NewFakeStorageBlobGetter(fakeBlob, test.blobs)
		finder := Finder{mockStorageBlobGetter, fakeNsgName}

		got, err := finder.FindSpecific(test.start, test.end)
		if test.err == nil && err != nil {
			t.Fatalf("unexpected error in test: %v", err)
		} else if test.err != nil && !errors.Is(err, test.err) {
			t.Fatalf("expected error: %v, got: %v", test.err, err)
		}

		gotPaths := make([]string, 0)
		for _, g := range got {
			gotPaths = append(gotPaths, g.Path)
		}

		sort.Strings(test.expected)
		sort.Strings(gotPaths)

		if !reflect.DeepEqual(test.expected, gotPaths) {
			t.Fatalf("expected: %v, got %v", test.expected, gotPaths)
		}
	}
}
