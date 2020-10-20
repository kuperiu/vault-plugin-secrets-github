package github

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/jsonutil"
	"github.com/hashicorp/vault/sdk/logical"
)

// Factory configures and returns Mock backends
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := &backend{
		store: make(map[string][]byte),
	}

	b.Backend = &framework.Backend{
		Help:        strings.TrimSpace(githubHelp),
		BackendType: logical.TypeLogical,
	}

	b.Backend.Paths = append(b.Backend.Paths, b.paths()...)

	if conf == nil {
		return nil, fmt.Errorf("configuration passed into backend is nil")
	}

	b.Backend.Setup(ctx, conf)

	return b, nil
}

// backend wraps the backend framework and adds a map for storing key value pairs
type backend struct {
	*framework.Backend

	store map[string][]byte
}

func (b *backend) paths() []*framework.Path {
	return []*framework.Path{
		{
			Pattern: framework.MatchAllRegex("path"),

			Fields: map[string]*framework.FieldSchema{
				"path": {
					Type:        framework.TypeString,
					Description: "Specifies the path of the secret.",
				},
			},

			Operations: map[logical.Operation]framework.OperationHandler{
				logical.ReadOperation: &framework.PathOperation{
					Callback: b.handleRead,
					Summary:  "Retrieve the secret from the map.",
				},
				logical.UpdateOperation: &framework.PathOperation{
					Callback: b.handleWrite,
					Summary:  "Store a secret at the specified location.",
				},
				logical.CreateOperation: &framework.PathOperation{
					Callback: b.handleWrite,
				},
				logical.DeleteOperation: &framework.PathOperation{
					Callback: b.handleDelete,
					Summary:  "Deletes the secret at the specified location.",
				},
			},

			ExistenceCheck: b.handleExistenceCheck,
		},
	}
}

func (b *backend) handleExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	out, err := req.Storage.Get(ctx, req.Path)
	if err != nil {
		return false, errwrap.Wrapf("existence check failed: {{err}}", err)
	}

	return out != nil, nil
}

func (b *backend) handleRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if req.ClientToken == "" {
		return nil, fmt.Errorf("client token empty")
	}

	path := data.Get("path").(string)

	// Decode the data
	var rawData map[string]interface{}
	if err := jsonutil.DecodeJSON(b.store[req.ClientToken+"/"+path], &rawData); err != nil {
		return nil, errwrap.Wrapf("json decoding failed: {{err}}", err)
	}

	appID, _ := rawData["app_id"].(json.Number).Int64()
	installationID := rawData["installation_id"].(json.Number).String()
	key := []byte(rawData["key"].(string))
	accessToken, err := getInstallationToken(key, appID, installationID)
	if err != nil {
		return nil, err
	}
	t := structToMapViaReflect(accessToken)

	// Generate the response
	resp := &logical.Response{
		Data: t,
	}

	return resp, nil
}

func (b *backend) handleWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	var buf []byte
	if req.ClientToken == "" {
		return nil, fmt.Errorf("client token empty")
	}

	// Check to make sure that kv pairs provided
	if len(req.Data) == 0 {
		return nil, fmt.Errorf("data must be provided to store in secret")
	}

	path := data.Get("path").(string)

	err := checkKeys(req.Data)
	if err != nil {
		return nil, err
	} else {
		// JSON encode the data
		buf, err = json.Marshal(req.Data)
		if err != nil {
			return nil, errwrap.Wrapf("json encoding failed: {{err}}", err)
		}
	}

	// Store kv pairs in map at specified path
	b.store[req.ClientToken+"/"+path] = buf

	return nil, nil
}

func (b *backend) handleDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if req.ClientToken == "" {
		return nil, fmt.Errorf("client token empty")
	}

	path := data.Get("path").(string)

	// Remove entry for specified path
	delete(b.store, path)

	return nil, nil
}

const githubHelp = `
The GitHUB backend is a secret engine for GitHub apps private keys that genrates installation tokens.
`

func checkKeys(m map[string]interface{}) error {
	keys := []string{"key", "app_id", "installation_id"}
	for i := range keys {
		_, ok := m[keys[i]]
		if !ok {
			return fmt.Errorf("%s is not in payload", keys[i])
		}
	}
	return nil
}

func structToMapViaReflect(accessToken AccessToken) map[string]interface{} {
	m := make(map[string]interface{})
	elem := reflect.ValueOf(&accessToken).Elem()
	relType := elem.Type()
	for i := 0; i < relType.NumField(); i++ {
		m[relType.Field(i).Name] = elem.Field(i).Interface()
	}
	return m
}
