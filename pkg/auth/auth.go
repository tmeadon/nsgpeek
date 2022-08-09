package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

func GetCredential() (azcore.TokenCredential, error) {
	cred, err := azidentity.NewAzureCLICredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a credential: %w", err)
	}
	return cred, nil
}

func GetSubscriptions(cred azcore.TokenCredential) ([]string, error) {
	t, err := getToken(cred)
	if err != nil {
		return nil, err
	}

	s, err := listSubscriptions(t)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	return s, nil
}

func getToken(cred azcore.TokenCredential) (string, error) {
	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{"https://management.azure.com/"},
	})

	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	return token.Token, nil
}

func listSubscriptions(token string) ([]string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://management.azure.com/subscriptions?api-version=2022-01-01", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call subscriptions API: %w", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var subList subscriptionList
	err = json.Unmarshal(body, &subList)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialise subscription response: %w", err)
	}

	subs := make([]string, 0)

	for _, s := range subList.value {
		subs = append(subs, s.subscriptionId)
	}

	return subs, nil
}

type subscriptionList struct {
	value []subscription
}

type subscription struct {
	subscriptionId string
}
