package centraldogma

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

type appIdentityService service

// AppIdentityType represents a type of app identity.
type AppIdentityType string

const (
	AppIdentityTypeToken       AppIdentityType = "TOKEN"
	AppIdentityTypeCertificate AppIdentityType = "CERTIFICATE"
)

// ProjectRole represents a role of an app identity in a project.
type ProjectRole string

const (
	ProjectRoleOwner  ProjectRole = "OWNER"
	ProjectRoleMember ProjectRole = "MEMBER"
)

// UserAndTimestamp represents who performed an action and when.
type UserAndTimestamp struct {
	User      string `json:"user,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// AppIdentity represents an app identity in the Central Dogma server.
type AppIdentity struct {
	AppID            string            `json:"appId"`
	Type             AppIdentityType   `json:"type,omitempty"`
	Secret           string            `json:"secret,omitempty"`
	CertificateID    string            `json:"certificateId,omitempty"`
	SystemAdmin      bool              `json:"systemAdmin,omitempty"`
	AllowGuestAccess bool              `json:"allowGuestAccess,omitempty"`
	Creation         *UserAndTimestamp `json:"creation,omitempty"`
	Deactivation     *UserAndTimestamp `json:"deactivation,omitempty"`
	Deletion         *UserAndTimestamp `json:"deletion,omitempty"`
}

// CreateAppIdentityRequest is used for creating an app identity.
type CreateAppIdentityRequest struct {
	AppID         string
	Type          AppIdentityType
	IsSystemAdmin bool
	Secret        string
	CertificateID string
}

func (a *appIdentityService) create(ctx context.Context, request *CreateAppIdentityRequest) (*AppIdentity, int, error) {
	u, err := url.Parse(path.Join(defaultPathPrefix, appIdentities))
	if err != nil {
		return nil, UnknownHttpStatusCode, err
	}

	q := u.Query()
	q.Set("appId", request.AppID)
	q.Set("type", string(request.Type))
	q.Set("isSystemAdmin", strconv.FormatBool(request.IsSystemAdmin))
	if request.Secret != "" {
		q.Set("secret", request.Secret)
	}
	if request.CertificateID != "" {
		q.Set("certificateId", request.CertificateID)
	}
	u.RawQuery = q.Encode()

	req, err := a.client.newRequest(http.MethodPost, u, nil)
	if err != nil {
		return nil, UnknownHttpStatusCode, err
	}

	appIdentity := new(AppIdentity)
	httpStatusCode, err := a.client.do(ctx, req, appIdentity, false)
	if err != nil {
		return nil, httpStatusCode, err
	}

	return appIdentity, httpStatusCode, nil
}

func (a *appIdentityService) addToProject(
	ctx context.Context, projectName, appID string, role ProjectRole) (int, int, error) {
	u, err := url.Parse(path.Join(
		defaultPathPrefix,
		metadata, projectName,
		appIdentities,
	))
	if err != nil {
		return -1, UnknownHttpStatusCode, err
	}

	body := map[string]string{"id": appID, "role": string(role)}
	req, err := a.client.newRequest(http.MethodPost, u, body)
	if err != nil {
		return -1, UnknownHttpStatusCode, err
	}

	revision := new(rev)
	httpStatusCode, err := a.client.do(ctx, req, revision, false)
	if err != nil {
		return -1, httpStatusCode, err
	}

	return revision.Rev, httpStatusCode, nil
}
