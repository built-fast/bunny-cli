package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/built-fast/bunny-cli/internal/client"
)

// --- dns records list ---

func TestDnsRecordsList_Table(t *testing.T) {
	t.Parallel()
	mock := &mockDnsZoneAPI{
		getDnsZoneFn: func(_ context.Context, id int64) (*client.DnsZone, error) {
			if id != 100 {
				t.Errorf("expected id=100, got %d", id)
			}
			return sampleDnsZone(), nil
		},
	}
	app := newTestDnsZoneApp(mock)

	out, _, err := executeCommand(app, "dns", "records", "list", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "93.184.216.34") {
		t.Error("expected output to contain record value")
	}
	if !strings.Contains(out, "A") {
		t.Error("expected output to contain record type name")
	}
	if !strings.Contains(out, "CNAME") {
		t.Error("expected output to contain CNAME record type")
	}
}

func TestDnsRecordsList_WithoutZoneID_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "dns", "records", "list")
	if err == nil {
		t.Fatal("expected error for missing zone ID argument")
	}
}

// --- dns records add ---

func TestDnsRecordsAdd_Success(t *testing.T) {
	t.Parallel()
	var capturedZoneId int64
	var capturedBody *client.DnsRecordCreate
	mock := &mockDnsZoneAPI{
		addDnsRecordFn: func(_ context.Context, zoneId int64, body *client.DnsRecordCreate) (*client.DnsRecord, error) {
			capturedZoneId = zoneId
			capturedBody = body
			return &client.DnsRecord{Id: 10, Type: body.Type, Name: body.Name, Value: body.Value, Ttl: body.Ttl}, nil
		},
	}
	app := newTestDnsZoneApp(mock)

	out, _, err := executeCommand(app, "dns", "records", "add", "100", "--type", "A", "--value", "1.2.3.4", "--name", "www", "--ttl", "300")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedZoneId != 100 {
		t.Errorf("expected zoneId=100, got %d", capturedZoneId)
	}
	if capturedBody.Type != 0 { // A = 0
		t.Errorf("expected type=0 (A), got %d", capturedBody.Type)
	}
	if capturedBody.Value != "1.2.3.4" {
		t.Errorf("expected value='1.2.3.4', got %q", capturedBody.Value)
	}
	if capturedBody.Name != "www" {
		t.Errorf("expected name='www', got %q", capturedBody.Name)
	}
	if !strings.Contains(out, "1.2.3.4") {
		t.Error("expected output to contain record value")
	}
}

func TestDnsRecordsAdd_TypeCaseInsensitive(t *testing.T) {
	t.Parallel()
	var capturedBody *client.DnsRecordCreate
	mock := &mockDnsZoneAPI{
		addDnsRecordFn: func(_ context.Context, zoneId int64, body *client.DnsRecordCreate) (*client.DnsRecord, error) {
			capturedBody = body
			return &client.DnsRecord{Id: 10, Type: body.Type, Value: body.Value}, nil
		},
	}
	app := newTestDnsZoneApp(mock)

	_, _, err := executeCommand(app, "dns", "records", "add", "100", "--type", "cname", "--value", "example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody.Type != 2 { // CNAME = 2
		t.Errorf("expected type=2 (CNAME), got %d", capturedBody.Type)
	}
}

func TestDnsRecordsAdd_InvalidType(t *testing.T) {
	t.Parallel()
	mock := &mockDnsZoneAPI{
		addDnsRecordFn: func(_ context.Context, zoneId int64, body *client.DnsRecordCreate) (*client.DnsRecord, error) {
			return sampleDnsRecord(), nil
		},
	}
	app := newTestDnsZoneApp(mock)

	_, stderr, err := executeCommand(app, "dns", "records", "add", "100", "--type", "INVALID", "--value", "test")
	if err == nil {
		t.Fatal("expected error for invalid record type")
	}
	if !strings.Contains(stderr, "unknown DNS record type") {
		t.Errorf("expected 'unknown DNS record type' error, got %q", stderr)
	}
}

func TestDnsRecordsAdd_FromFile(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "dns", "records", "add", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--from-file") {
		t.Error("expected add command to have --from-file flag")
	}
}

// --- dns records update ---

func TestDnsRecordsUpdate_Success(t *testing.T) {
	t.Parallel()
	var capturedZoneId, capturedRecordId int64
	var capturedBody *client.DnsRecordUpdate
	mock := &mockDnsZoneAPI{
		updateDnsRecordFn: func(_ context.Context, zoneId, recordId int64, body *client.DnsRecordUpdate) error {
			capturedZoneId = zoneId
			capturedRecordId = recordId
			capturedBody = body
			return nil
		},
	}
	app := newTestDnsZoneApp(mock)

	out, _, err := executeCommand(app, "dns", "records", "update", "100", "1", "--value", "5.6.7.8", "--ttl", "600")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedZoneId != 100 {
		t.Errorf("expected zoneId=100, got %d", capturedZoneId)
	}
	if capturedRecordId != 1 {
		t.Errorf("expected recordId=1, got %d", capturedRecordId)
	}
	if capturedBody.Value == nil || *capturedBody.Value != "5.6.7.8" {
		t.Error("expected value to be set")
	}
	if capturedBody.Ttl == nil || *capturedBody.Ttl != 600 {
		t.Error("expected ttl to be set")
	}
	if !strings.Contains(out, "DNS record updated") {
		t.Error("expected update confirmation message")
	}
}

func TestDnsRecordsUpdate_WithoutArgs_Fails(t *testing.T) {
	t.Parallel()
	_, _, err := executeCommand(nil, "dns", "records", "update")
	if err == nil {
		t.Fatal("expected error for missing arguments")
	}
}

// --- dns records delete ---

func TestDnsRecordsDelete_WithYes(t *testing.T) {
	t.Parallel()
	var capturedZoneId, capturedRecordId int64
	mock := &mockDnsZoneAPI{
		deleteDnsRecordFn: func(_ context.Context, zoneId, recordId int64) error {
			capturedZoneId = zoneId
			capturedRecordId = recordId
			return nil
		},
	}
	app := newTestDnsZoneApp(mock)

	out, _, err := executeCommand(app, "dns", "records", "delete", "100", "1", "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedZoneId != 100 {
		t.Errorf("expected zoneId=100, got %d", capturedZoneId)
	}
	if capturedRecordId != 1 {
		t.Errorf("expected recordId=1, got %d", capturedRecordId)
	}
	if !strings.Contains(out, "DNS record deleted") {
		t.Error("expected deletion confirmation message")
	}
}

func TestDnsRecordsDelete_Canceled(t *testing.T) {
	t.Parallel()
	mock := &mockDnsZoneAPI{
		deleteDnsRecordFn: func(_ context.Context, zoneId, recordId int64) error {
			t.Error("delete should not have been called")
			return nil
		},
	}
	app := newTestDnsZoneApp(mock)

	stdin := bytes.NewBufferString("n\n")
	_, stderr, err := executeCommandWithStdin(app, stdin, "dns", "records", "delete", "100", "1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stderr, "Deletion canceled") {
		t.Error("expected cancellation message")
	}
}
