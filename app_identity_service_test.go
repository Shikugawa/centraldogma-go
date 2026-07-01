package centraldogma

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestCreateAppIdentity(t *testing.T) {
	c, mux, teardown := setup()
	defer teardown()

	request := &CreateAppIdentityRequest{
		AppID:         "my-app",
		Type:          AppIdentityTypeToken,
		IsSystemAdmin: false,
	}

	mux.HandleFunc("/api/v1/appIdentities", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testAuthorization(t, r)
		testURLQuery(t, r, "appId", "my-app")
		testURLQuery(t, r, "type", "TOKEN")
		testURLQuery(t, r, "isSystemAdmin", "false")
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Location", "/api/v1/appIdentities/my-app")
		fmt.Fprint(w, `{"appId":"my-app","type":"TOKEN","secret":"appToken-secret","systemAdmin":false}`)
	})

	appIdentity, httpStatusCode, _ := c.CreateAppIdentity(context.Background(), request)
	testStatusCode(t, httpStatusCode, 201)

	want := &AppIdentity{AppID: "my-app", Type: AppIdentityTypeToken, Secret: "appToken-secret"}
	if !reflect.DeepEqual(appIdentity, want) {
		t.Errorf("CreateAppIdentity returned %+v, want %+v", appIdentity, want)
	}
}

func TestCreateCertificateAppIdentity(t *testing.T) {
	c, mux, teardown := setup()
	defer teardown()

	request := &CreateAppIdentityRequest{
		AppID:         "mtls-app",
		Type:          AppIdentityTypeCertificate,
		IsSystemAdmin: true,
		CertificateID: "cert-1",
	}

	mux.HandleFunc("/api/v1/appIdentities", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testURLQuery(t, r, "appId", "mtls-app")
		testURLQuery(t, r, "type", "CERTIFICATE")
		testURLQuery(t, r, "isSystemAdmin", "true")
		testURLQuery(t, r, "certificateId", "cert-1")
		fmt.Fprint(w, `{"appId":"mtls-app","type":"CERTIFICATE","certificateId":"cert-1","systemAdmin":true}`)
	})

	appIdentity, _, _ := c.CreateAppIdentity(context.Background(), request)
	want := &AppIdentity{AppID: "mtls-app", Type: AppIdentityTypeCertificate, CertificateID: "cert-1", SystemAdmin: true}
	if !reflect.DeepEqual(appIdentity, want) {
		t.Errorf("CreateAppIdentity returned %+v, want %+v", appIdentity, want)
	}
}

func TestAddAppIdentityToProject(t *testing.T) {
	c, mux, teardown := setup()
	defer teardown()

	input := map[string]string{"id": "my-app", "role": "MEMBER"}

	mux.HandleFunc("/api/v1/metadata/foo/appIdentities", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testAuthorization(t, r)

		body := map[string]string{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if !reflect.DeepEqual(body, input) {
			t.Errorf("Request body = %+v, want %+v", body, input)
		}

		fmt.Fprint(w, `{"revision":7}`)
	})

	revision, httpStatusCode, _ := c.AddAppIdentityToProject(context.Background(), "foo", "my-app", ProjectRoleMember)
	testStatusCode(t, httpStatusCode, 200)
	if revision != 7 {
		t.Errorf("AddAppIdentityToProject returned %v, want %v", revision, 7)
	}
}
