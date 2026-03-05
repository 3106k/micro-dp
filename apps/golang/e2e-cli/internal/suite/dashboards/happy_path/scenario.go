package happy_path

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/user/micro-dp/e2e-cli/internal/httpclient"
	"github.com/user/micro-dp/e2e-cli/internal/openapi"
)

type Scenario struct {
	password    string
	displayName string
}

func NewScenario(password, displayName string) *Scenario {
	return &Scenario{
		password:    password,
		displayName: displayName,
	}
}

func (s *Scenario) ID() string {
	return "dashboards/happy_path/crud_with_charts_and_widgets"
}

func (s *Scenario) Run(ctx context.Context, client *httpclient.Client) error {
	ts := time.Now().UnixNano()

	// === Auth setup ===
	email := fmt.Sprintf("e2e_dashboards_%d@example.com", ts)
	registerReq := openapi.RegisterRequest{
		Email:       openapi.Email(email),
		Password:    s.password,
		DisplayName: openapi.Ptr(s.displayName),
	}
	var registerResp openapi.RegisterResponse
	code, body, err := client.PostJSON(ctx, "/api/v1/auth/register", registerReq, &registerResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("register: status=%d body=%s", code, string(body))
	}

	loginReq := openapi.LoginRequest{
		Email:    openapi.Email(email),
		Password: s.password,
	}
	var loginResp openapi.LoginResponse
	code, body, err = client.PostJSON(ctx, "/api/v1/auth/login", loginReq, &loginResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("login: status=%d body=%s", code, string(body))
	}
	client.SetToken(loginResp.Token)
	client.SetTenantID(registerResp.TenantId)

	// === Upload CSV to create a dataset ===
	datasetID, err := s.uploadCSVAndWaitForDataset(ctx, client)
	if err != nil {
		return fmt.Errorf("dataset setup: %w", err)
	}

	// === Chart CRUD ===
	chartID, err := s.testChartCRUD(ctx, client, datasetID)
	if err != nil {
		return err
	}

	// === Chart Data ===
	if err := s.testChartData(ctx, client, chartID); err != nil {
		return err
	}

	// === Dashboard CRUD ===
	dashboardID, err := s.testDashboardCRUD(ctx, client)
	if err != nil {
		return err
	}

	// === Widget CRUD ===
	if err := s.testWidgetCRUD(ctx, client, dashboardID, chartID); err != nil {
		return err
	}

	// === Cleanup: delete chart and dashboard ===
	code, body, err = client.Delete(ctx, "/api/v1/charts/"+chartID)
	if err != nil {
		return err
	}
	if code != 204 {
		return fmt.Errorf("delete chart: expected 204 got=%d body=%s", code, string(body))
	}

	code, body, err = client.Delete(ctx, "/api/v1/dashboards/"+dashboardID)
	if err != nil {
		return err
	}
	if code != 204 {
		return fmt.Errorf("delete dashboard: expected 204 got=%d body=%s", code, string(body))
	}

	return nil
}

func (s *Scenario) uploadCSVAndWaitForDataset(ctx context.Context, client *httpclient.Client) (string, error) {
	// Presign
	presignReq := openapi.CreateUploadPresignRequest{
		Files: []openapi.UploadFileInput{
			{Filename: "e2e-chart-data.csv", ContentType: "text/csv", SizeBytes: 1024},
		},
	}
	var presignResp openapi.CreateUploadPresignResponse
	code, body, err := client.PostJSON(ctx, "/api/v1/uploads/presign", presignReq, &presignResp)
	if err != nil {
		return "", err
	}
	if code != 201 {
		return "", fmt.Errorf("presign: status=%d body=%s", code, string(body))
	}

	// Upload CSV
	csvData := []byte("category,amount\nA,100\nB,200\nC,150\nD,300\n")
	putReq, err := http.NewRequestWithContext(ctx, http.MethodPut, presignResp.Files[0].PresignedUrl, bytes.NewReader(csvData))
	if err != nil {
		return "", fmt.Errorf("create put request: %w", err)
	}
	putReq.Header.Set("Content-Type", "text/csv")
	putResp, err := http.DefaultClient.Do(putReq)
	if err != nil {
		return "", fmt.Errorf("put csv: %w", err)
	}
	putResp.Body.Close()
	if putResp.StatusCode != 200 {
		return "", fmt.Errorf("put csv: expected 200 got=%d", putResp.StatusCode)
	}

	// Complete
	var completeResp openapi.Upload
	code, body, err = client.PostJSON(ctx, "/api/v1/uploads/"+presignResp.UploadId+"/complete", nil, &completeResp)
	if err != nil {
		return "", err
	}
	if code != 200 {
		return "", fmt.Errorf("complete: status=%d body=%s", code, string(body))
	}

	// Poll for dataset
	for attempt := 0; attempt < 15; attempt++ {
		time.Sleep(1 * time.Second)
		var datasetsResp openapi.ListResponse[openapi.Dataset]
		code, body, err = client.GetJSON(ctx, "/api/v1/datasets?source_type=import", &datasetsResp)
		if err != nil {
			return "", err
		}
		if code != 200 {
			return "", fmt.Errorf("datasets list: status=%d body=%s", code, string(body))
		}
		for _, ds := range datasetsResp.Items {
			if ds.Name == "e2e-chart-data" {
				return ds.Id, nil
			}
		}
	}
	return "", fmt.Errorf("dataset 'e2e-chart-data' not found after 15s")
}

func (s *Scenario) testChartCRUD(ctx context.Context, client *httpclient.Client, datasetID string) (string, error) {
	// List charts → empty
	var listResp openapi.ListResponse[openapi.Chart]
	code, body, err := client.GetJSON(ctx, "/api/v1/charts", &listResp)
	if err != nil {
		return "", err
	}
	if code != 200 {
		return "", fmt.Errorf("list charts: status=%d body=%s", code, string(body))
	}
	if len(listResp.Items) != 0 {
		return "", fmt.Errorf("list charts: expected 0 items, got=%d", len(listResp.Items))
	}

	// Create chart
	createReq := openapi.CreateChartRequest{
		Name:      "E2E Bar Chart",
		ChartType: openapi.Bar,
		DatasetId: datasetID,
		Measure:   "amount",
		Dimension: "category",
	}
	var createResp openapi.Chart
	code, body, err = client.PostJSON(ctx, "/api/v1/charts", createReq, &createResp)
	if err != nil {
		return "", err
	}
	if code != 201 {
		return "", fmt.Errorf("create chart: status=%d body=%s", code, string(body))
	}
	if createResp.Id == "" {
		return "", fmt.Errorf("create chart: id is empty")
	}
	chartID := createResp.Id

	// Get chart
	var getResp openapi.Chart
	code, body, err = client.GetJSON(ctx, "/api/v1/charts/"+chartID, &getResp)
	if err != nil {
		return "", err
	}
	if code != 200 {
		return "", fmt.Errorf("get chart: status=%d body=%s", code, string(body))
	}
	if getResp.Name != "E2E Bar Chart" {
		return "", fmt.Errorf("get chart: expected name='E2E Bar Chart' got='%s'", getResp.Name)
	}
	if getResp.ChartType != openapi.Bar {
		return "", fmt.Errorf("get chart: expected type=bar got=%s", getResp.ChartType)
	}
	if getResp.Measure != "amount" {
		return "", fmt.Errorf("get chart: expected measure=amount got=%s", getResp.Measure)
	}
	if getResp.Dimension != "category" {
		return "", fmt.Errorf("get chart: expected dimension=category got=%s", getResp.Dimension)
	}

	// Update chart
	updateReq := openapi.UpdateChartRequest{
		Name:      "E2E Line Chart",
		ChartType: openapi.Line,
		DatasetId: datasetID,
		Measure:   "amount",
		Dimension: "category",
	}
	var updateResp openapi.Chart
	code, body, err = client.PutJSON(ctx, "/api/v1/charts/"+chartID, updateReq, &updateResp)
	if err != nil {
		return "", err
	}
	if code != 200 {
		return "", fmt.Errorf("update chart: status=%d body=%s", code, string(body))
	}
	if updateResp.Name != "E2E Line Chart" {
		return "", fmt.Errorf("update chart: expected name='E2E Line Chart' got='%s'", updateResp.Name)
	}
	if updateResp.ChartType != openapi.Line {
		return "", fmt.Errorf("update chart: expected type=line got=%s", updateResp.ChartType)
	}

	// List charts → 1
	code, body, err = client.GetJSON(ctx, "/api/v1/charts", &listResp)
	if err != nil {
		return "", err
	}
	if code != 200 {
		return "", fmt.Errorf("list charts after create: status=%d body=%s", code, string(body))
	}
	if len(listResp.Items) != 1 {
		return "", fmt.Errorf("list charts: expected 1 item, got=%d", len(listResp.Items))
	}

	return chartID, nil
}

func (s *Scenario) testChartData(ctx context.Context, client *httpclient.Client, chartID string) error {
	var dataResp openapi.ChartDataResponse
	code, body, err := client.GetJSON(ctx, "/api/v1/charts/"+chartID+"/data?period=last_30_days", &dataResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("chart data: status=%d body=%s", code, string(body))
	}
	if dataResp.ChartId != chartID {
		return fmt.Errorf("chart data: expected chart_id=%s got=%s", chartID, dataResp.ChartId)
	}
	if len(dataResp.Labels) == 0 {
		return fmt.Errorf("chart data: labels is empty")
	}
	if len(dataResp.Datasets) == 0 {
		return fmt.Errorf("chart data: datasets is empty")
	}
	if len(dataResp.Datasets[0].Data) == 0 {
		return fmt.Errorf("chart data: datasets[0].data is empty")
	}

	// Verify nonexistent chart → 404
	code, body, err = client.GetJSON(ctx, "/api/v1/charts/nonexistent/data?period=last_30_days", nil)
	if err != nil {
		return err
	}
	if code != 404 {
		return fmt.Errorf("chart data nonexistent: expected 404 got=%d body=%s", code, string(body))
	}

	return nil
}

func (s *Scenario) testDashboardCRUD(ctx context.Context, client *httpclient.Client) (string, error) {
	// List dashboards → empty
	var listResp openapi.ListResponse[openapi.Dashboard]
	code, body, err := client.GetJSON(ctx, "/api/v1/dashboards", &listResp)
	if err != nil {
		return "", err
	}
	if code != 200 {
		return "", fmt.Errorf("list dashboards: status=%d body=%s", code, string(body))
	}
	if len(listResp.Items) != 0 {
		return "", fmt.Errorf("list dashboards: expected 0 items, got=%d", len(listResp.Items))
	}

	// Create dashboard
	createReq := openapi.CreateDashboardRequest{
		Name:        "E2E Dashboard",
		Description: openapi.Ptr("E2E test dashboard"),
	}
	var createResp openapi.Dashboard
	code, body, err = client.PostJSON(ctx, "/api/v1/dashboards", createReq, &createResp)
	if err != nil {
		return "", err
	}
	if code != 201 {
		return "", fmt.Errorf("create dashboard: status=%d body=%s", code, string(body))
	}
	if createResp.Id == "" {
		return "", fmt.Errorf("create dashboard: id is empty")
	}
	dashID := createResp.Id

	// Get dashboard
	var getResp openapi.Dashboard
	code, body, err = client.GetJSON(ctx, "/api/v1/dashboards/"+dashID, &getResp)
	if err != nil {
		return "", err
	}
	if code != 200 {
		return "", fmt.Errorf("get dashboard: status=%d body=%s", code, string(body))
	}
	if getResp.Name != "E2E Dashboard" {
		return "", fmt.Errorf("get dashboard: expected name='E2E Dashboard' got='%s'", getResp.Name)
	}
	if getResp.Description == nil || *getResp.Description != "E2E test dashboard" {
		return "", fmt.Errorf("get dashboard: description mismatch")
	}

	// Update dashboard
	updateReq := openapi.UpdateDashboardRequest{
		Name:        "E2E Dashboard Updated",
		Description: openapi.Ptr("Updated description"),
	}
	var updateResp openapi.Dashboard
	code, body, err = client.PutJSON(ctx, "/api/v1/dashboards/"+dashID, updateReq, &updateResp)
	if err != nil {
		return "", err
	}
	if code != 200 {
		return "", fmt.Errorf("update dashboard: status=%d body=%s", code, string(body))
	}
	if updateResp.Name != "E2E Dashboard Updated" {
		return "", fmt.Errorf("update dashboard: expected name='E2E Dashboard Updated' got='%s'", updateResp.Name)
	}

	// List dashboards → 1
	code, body, err = client.GetJSON(ctx, "/api/v1/dashboards", &listResp)
	if err != nil {
		return "", err
	}
	if code != 200 {
		return "", fmt.Errorf("list dashboards after create: status=%d body=%s", code, string(body))
	}
	if len(listResp.Items) != 1 {
		return "", fmt.Errorf("list dashboards: expected 1 item, got=%d", len(listResp.Items))
	}

	return dashID, nil
}

func (s *Scenario) testWidgetCRUD(ctx context.Context, client *httpclient.Client, dashboardID, chartID string) error {
	basePath := "/api/v1/dashboards/" + dashboardID + "/widgets"

	// List widgets → empty
	var listResp openapi.ListResponse[openapi.DashboardWidget]
	code, body, err := client.GetJSON(ctx, basePath, &listResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("list widgets: status=%d body=%s", code, string(body))
	}
	if len(listResp.Items) != 0 {
		return fmt.Errorf("list widgets: expected 0 items, got=%d", len(listResp.Items))
	}

	// Create widget
	createReq := openapi.CreateDashboardWidgetRequest{
		ChartId:  chartID,
		Position: 0,
	}
	var createResp openapi.DashboardWidget
	code, body, err = client.PostJSON(ctx, basePath, createReq, &createResp)
	if err != nil {
		return err
	}
	if code != 201 {
		return fmt.Errorf("create widget: status=%d body=%s", code, string(body))
	}
	if createResp.Id == "" {
		return fmt.Errorf("create widget: id is empty")
	}
	widgetID := createResp.Id

	// List widgets → 1
	code, body, err = client.GetJSON(ctx, basePath, &listResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("list widgets after create: status=%d body=%s", code, string(body))
	}
	if len(listResp.Items) != 1 {
		return fmt.Errorf("list widgets: expected 1 item, got=%d", len(listResp.Items))
	}
	if listResp.Items[0].ChartId != chartID {
		return fmt.Errorf("widget chart_id: expected %s got=%s", chartID, listResp.Items[0].ChartId)
	}

	// Delete widget
	code, body, err = client.Delete(ctx, basePath+"/"+widgetID)
	if err != nil {
		return err
	}
	if code != 204 {
		return fmt.Errorf("delete widget: expected 204 got=%d body=%s", code, string(body))
	}

	// List widgets → empty again
	code, body, err = client.GetJSON(ctx, basePath, &listResp)
	if err != nil {
		return err
	}
	if code != 200 {
		return fmt.Errorf("list widgets after delete: status=%d body=%s", code, string(body))
	}
	if len(listResp.Items) != 0 {
		return fmt.Errorf("list widgets after delete: expected 0 items, got=%d", len(listResp.Items))
	}

	return nil
}
