package flowwriter

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

var testFlows string = `
{
  "time": "2022-08-09T10:03:27.7257644Z",
  "systemId": "e79aab03-ffb0-4419-8a28-90be262a7028",
  "macAddress": "000D3AD488D1",
  "category": "NetworkSecurityGroupFlowEvent",
  "resourceId": "/SUBSCRIPTIONS/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx/RESOURCEGROUPS/NSG-VIEW/PROVIDERS/MICROSOFT.NETWORK/NETWORKSECURITYGROUPS/NSG-VIEW",
  "operationName": "NetworkSecurityGroupFlowEvents",
  "properties": {
    "Version": 2,
    "flows": [
      {
        "rule": "DefaultRule_AllowInternetOutBound",
        "flows": [
          {
            "mac": "000D3AD488D1",
            "flowTuples": [
              "1660039344,10.0.0.4,51.104.229.52,50276,443,T,O,A,E,14,2839,14,5801",
              "1660039344,10.0.0.4,51.105.74.153,47382,443,T,O,B,B,,,,",
              "1660039350,10.0.0.4,51.105.74.153,47382,443,T,O,A,C,12,3769,10,5061"
            ]
          }
        ]
      },
      {
        "rule": "DefaultRule_DenyAllInBound",
        "flows": [
          {
            "mac": "000D3AD488D1",
            "flowTuples": [
              "1660039356,117.88.229.255,10.0.0.4,50996,23,T,I,D,B,,,,",
              "1660039361,167.99.14.84,10.0.0.4,39984,8080,T,I,D,B,,,,",
              "1660039369,176.63.187.19,10.0.0.4,46852,23,T,I,D,B,,,,"
            ]
          }
        ]
      },
      {
        "rule": "UserRule_ssh",
        "flows": [
          {
            "mac": "000D3AD488D1",
            "flowTuples": [
              "1660039351,38.88.252.187,10.0.0.4,59246,22,T,I,A,B,,,,",
              "1660039358,61.177.173.21,10.0.0.4,56496,22,T,I,A,B,,,,"
            ]
          }
        ]
      }
    ]
  }
}
`

var wantedLines = [][]string{
	{"Aug", "9", "11:02:24.000", "DefaultRule_AllowInternetOutBound", "10.0.0.4", "50276", "51.104.229.52", "443", "out", "allow", "end", "2839", "5801"},
	{"Aug", "9", "11:02:24.000", "DefaultRule_AllowInternetOutBound", "10.0.0.4", "47382", "51.105.74.153", "443", "out", "deny", "begin"},
	{"Aug", "9", "11:02:30.000", "DefaultRule_AllowInternetOutBound", "10.0.0.4", "47382", "51.105.74.153", "443", "out", "allow", "continuing", "3769", "5061"},
	{"Aug", "9", "11:02:36.000", "DefaultRule_DenyAllInBound", "117.88.229.255", "50996", "10.0.0.4", "23", "in", "deny", "begin"},
	{"Aug", "9", "11:02:41.000", "DefaultRule_DenyAllInBound", "167.99.14.84", "39984", "10.0.0.4", "8080", "in", "deny", "begin"},
	{"Aug", "9", "11:02:49.000", "DefaultRule_DenyAllInBound", "176.63.187.19", "46852", "10.0.0.4", "23", "in", "deny", "begin"},
	{"Aug", "9", "11:02:31.000", "UserRule_ssh", "38.88.252.187", "59246", "10.0.0.4", "22", "in", "allow", "begin"},
	{"Aug", "9", "11:02:38.000", "UserRule_ssh", "61.177.173.21", "56496", "10.0.0.4", "22", "in", "allow", "begin"},
}

func TestConsoleWriter(t *testing.T) {
	var buffer bytes.Buffer
	testFlowWriter := NewConsoleWriter(&buffer)

	err := testFlowWriter.WriteFlowBlock([]byte(testFlows))
	if err != nil {
		log.Fatalf("failed to set up test: %v", err)
	}

	testFlowWriter.Flush()

	t.Run("TestHeader", func(t *testing.T) {
		headerLine := strings.Split(buffer.String(), "\n")[0]
		got := strings.Fields(headerLine)
		want := []string{"time", "rule", "src_addr", "src_port", "dst_addr", "dst_port", "direction", "decision", "state", "src_to_dst_bytes", "dst_to_src_bytes"}

		for i, h := range want {
			if i > (len(got)-1) || got[i] != h {
				t.Errorf("missing header '%v'", h)
			}
		}
	})

	t.Run("TestTableLines", func(t *testing.T) {
		allTableLines := strings.Split(buffer.String(), "\n")
		tableLines := allTableLines[2:(len(allTableLines) - 1)]

		for i, wanted := range wantedLines {
			got := strings.Fields(tableLines[i])

			for j, w := range wanted {
				if w != got[j] {
					t.Errorf("missing table column value: '%v' in line '%v'", w, tableLines[i])
				}
			}
		}
	})
}
