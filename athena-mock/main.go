package main

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	middleware2 "github.com/aws/smithy-go/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	uuid2 "github.com/google/uuid"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	log.Println("Starting Athena Mock....")
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/", HandleRequest)

	http.ListenAndServe(":4567", r)
}

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	result, err := io.ReadAll(r.Body)

	if err != nil {
		log.Printf("Couldn't read body %s", err)
		w.WriteHeader(500)
		return
	}

	target := r.Header.Get("X-Amz-Target")

	var body interface{}
	var status int

	switch target {
	case "AmazonAthena.StartQueryExecution":
		var startQueryRequest StartQueryExecutionInput
		err = json.Unmarshal(result, &startQueryRequest)
		if err != nil {
			status = 500
			break
		}

		body, status = startQuery(startQueryRequest)
	case "AmazonAthena.GetQueryExecution":
		var getQueryExecutionRequest GetQueryExecutionInput
		err = json.Unmarshal(result, &getQueryExecutionRequest)
		if err != nil {
			status = 500
			break
		}

		body, status = getQueryExecution(getQueryExecutionRequest)
	case "AmazonAthena.GetQueryResults":
		var getQueryResultsInput GetQueryResultsInput
		err = json.Unmarshal(result, &getQueryResultsInput)
		if err != nil {
			status = 500
			break
		}

		body, status = getQueryResults(getQueryResultsInput)
	default:
		log.Printf("\n\nUnknown request: %s:\n%s\n\n", target, result)
		status = 500
	}

	w.WriteHeader(status)
	if body != nil {
		jsonResponse, _ := json.Marshal(body)
		w.Write(jsonResponse)
	}
	return
}

var createTableRequests = make(map[string]bool)
var queryRequests = make(map[string]time.Time)

func startQuery(request StartQueryExecutionInput) (interface{}, int) {
	log.Printf("Got start query request: %+v", request)

	id := uuid2.New().String()
	if strings.Contains(*request.QueryString, "CREATE EXTERNAL TABLE") {
		createTableRequests[id] = true
	} else if strings.Contains(*request.QueryString, "SELECT ") {
		queryRequests[id] = time.Now()
	}

	return StartQueryExecutionOutput{QueryExecutionId: &id}, 200
}

func getQueryExecution(request GetQueryExecutionInput) (interface{}, int) {
	log.Printf("Got get query execution request: %+v: %s", request, *request.QueryExecutionId)

	state := types.QueryExecutionStateRunning
	var completionDateTime *int64
	if _, createTableRequestExists := createTableRequests[*request.QueryExecutionId]; createTableRequestExists {
		state = types.QueryExecutionStateSucceeded
		completionDateTime = IntPtr(time.Now().Add(time.Second * -5).Unix())
	} else if queryRequestedAt, queryRequestExists := queryRequests[*request.QueryExecutionId]; queryRequestExists {
		if queryRequestedAt.Add(time.Second * 5).Before(time.Now()) {
			state = types.QueryExecutionStateSucceeded
			completionDateTime = IntPtr(queryRequestedAt.Add(time.Second * 5).Unix())
		}
	} else {
		state = types.QueryExecutionStateFailed
	}

	return GetQueryExecutionOutput{
		QueryExecution: &QueryExecution{
			EngineVersion: &EngineVersion{
				EffectiveEngineVersion: StringPtr("mock"),
				SelectedEngineVersion:  StringPtr("mock"),
			},
			Query: StringPtr("abc"),
			QueryExecutionContext: &QueryExecutionContext{
				Catalog:  StringPtr("catalog"),
				Database: StringPtr("database"),
			},
			QueryExecutionId: request.QueryExecutionId,
			ResultConfiguration: &ResultConfiguration{
				AclConfiguration:        &AclConfiguration{},
				EncryptionConfiguration: nil,
				ExpectedBucketOwner:     nil,
				OutputLocation:          StringPtr("s3://somebucket"),
			},
			StatementType: "",
			Statistics:    nil,
			Status: &QueryExecutionStatus{
				AthenaError:        nil,
				CompletionDateTime: completionDateTime,
				State:              state,
				StateChangeReason:  nil,
				SubmissionDateTime: nil,
			},
			WorkGroup: StringPtr("workgroup"),
		},
	}, 200
}

func getQueryResults(request GetQueryResultsInput) (interface{}, int) {
	log.Printf("Got get query results request: %+v: %s", request, *request.QueryExecutionId)

	if _, createTableRequestExists := createTableRequests[*request.QueryExecutionId]; createTableRequestExists {
		return GetQueryResultsOutput{
			NextToken:             nil,
			ResultSet:             nil,
			UpdateCount:           IntPtr(0),
			ResultMetadata:        middleware2.Metadata{},
			noSmithyDocumentSerde: noSmithyDocumentSerde{},
		}, 200
	} else if queryRequestedAt, queryRequestExists := queryRequests[*request.QueryExecutionId]; queryRequestExists {
		if queryRequestedAt.Add(time.Second * 5).Before(time.Now()) {
			return GetQueryResultsOutput{
				NextToken: nil,
				ResultSet: &ResultSet{
					ResultSetMetadata: nil,
					Rows: []Row{
						{Data: []Datum{
							{VarCharValue: StringPtr("Product Name")},
							{VarCharValue: StringPtr("Value")},
						}},
						{Data: []Datum{
							{VarCharValue: StringPtr("Product A")},
							{VarCharValue: StringPtr("54")},
						}},
						{Data: []Datum{
							{VarCharValue: StringPtr("Product B")},
							{VarCharValue: StringPtr("122")},
						}},
					},
					noSmithyDocumentSerde: noSmithyDocumentSerde{},
				},
				UpdateCount:           IntPtr(0),
				ResultMetadata:        middleware2.Metadata{},
				noSmithyDocumentSerde: noSmithyDocumentSerde{},
			}, 200
		}
	}

	return nil, 500
}

func StringPtr(s string) *string {
	return &s
}

func IntPtr(i int64) *int64 {
	return &i
}

func TimePtr(t time.Time) *time.Time {
	return &t
}
