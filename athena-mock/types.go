package main

import (
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	smithydocument "github.com/aws/smithy-go/document"
	"github.com/aws/smithy-go/middleware"
)

type AclConfiguration struct {
	S3AclOption types.S3AclOption
	noSmithyDocumentSerde
}

type AthenaError struct {
	ErrorCategory *int32
	ErrorMessage  *string
	ErrorType     *int32
	Retryable     bool
	noSmithyDocumentSerde
}

type Column struct {
	Name    *string
	Comment *string
	Type    *string
	noSmithyDocumentSerde
}

type ColumnInfo struct {
	Name          *string
	Type          *string
	CaseSensitive bool
	CatalogName   *string
	Label         *string
	Nullable      types.ColumnNullable
	Precision     int32
	Scale         int32
	SchemaName    *string
	TableName     *string
	noSmithyDocumentSerde
}

type Database struct {
	Name        *string
	Description *string
	Parameters  map[string]string
	noSmithyDocumentSerde
}

type DataCatalog struct {
	Name        *string
	Type        types.DataCatalogType
	Description *string
	Parameters  map[string]string
	noSmithyDocumentSerde
}

type DataCatalogSummary struct {
	CatalogName *string
	Type        types.DataCatalogType
	noSmithyDocumentSerde
}

type Datum struct {
	VarCharValue *string
	noSmithyDocumentSerde
}

type EncryptionConfiguration struct {
	EncryptionOption types.EncryptionOption
	KmsKey           *string
	noSmithyDocumentSerde
}

type EngineVersion struct {
	EffectiveEngineVersion *string
	SelectedEngineVersion  *string
	noSmithyDocumentSerde
}

type NamedQuery struct {
	Database     *string
	Name         *string
	QueryString  *string
	Description  *string
	NamedQueryId *string
	WorkGroup    *string
	noSmithyDocumentSerde
}

type PreparedStatement struct {
	Description      *string
	LastModifiedTime *int64
	QueryStatement   *string
	StatementName    *string
	WorkGroupName    *string
	noSmithyDocumentSerde
}

type PreparedStatementSummary struct {
	LastModifiedTime *int64
	StatementName    *string
	noSmithyDocumentSerde
}

type QueryExecution struct {
	EngineVersion         *EngineVersion
	Query                 *string
	QueryExecutionContext *QueryExecutionContext
	QueryExecutionId      *string
	ResultConfiguration   *ResultConfiguration
	StatementType         types.StatementType
	Statistics            *QueryExecutionStatistics
	Status                *QueryExecutionStatus
	WorkGroup             *string
	noSmithyDocumentSerde
}

type QueryExecutionContext struct {
	Catalog  *string
	Database *string
	noSmithyDocumentSerde
}

type QueryExecutionStatistics struct {
	DataManifestLocation          *string
	DataScannedInBytes            *int64
	EngineExecutionTimeInMillis   *int64
	QueryPlanningTimeInMillis     *int64
	QueryQueueTimeInMillis        *int64
	ServiceProcessingTimeInMillis *int64
	TotalExecutionTimeInMillis    *int64
	noSmithyDocumentSerde
}

type QueryExecutionStatus struct {
	AthenaError        *AthenaError
	CompletionDateTime *int64
	State              types.QueryExecutionState
	StateChangeReason  *string
	SubmissionDateTime *int64
	noSmithyDocumentSerde
}

type ResultConfiguration struct {
	AclConfiguration        *AclConfiguration
	EncryptionConfiguration *EncryptionConfiguration
	ExpectedBucketOwner     *string
	OutputLocation          *string
	noSmithyDocumentSerde
}

type ResultConfigurationUpdates struct {
	AclConfiguration              *AclConfiguration
	EncryptionConfiguration       *EncryptionConfiguration
	ExpectedBucketOwner           *string
	OutputLocation                *string
	RemoveAclConfiguration        *bool
	RemoveEncryptionConfiguration *bool
	RemoveExpectedBucketOwner     *bool
	RemoveOutputLocation          *bool
	noSmithyDocumentSerde
}

type ResultSet struct {
	ResultSetMetadata *ResultSetMetadata
	Rows              []Row
	noSmithyDocumentSerde
}

type ResultSetMetadata struct {
	ColumnInfo []ColumnInfo
	noSmithyDocumentSerde
}

type Row struct {
	Data []Datum
	noSmithyDocumentSerde
}

type TableMetadata struct {
	Name           *string
	Columns        []Column
	CreateTime     *int64
	LastAccessTime *int64
	Parameters     map[string]string
	PartitionKeys  []Column
	TableType      *string
	noSmithyDocumentSerde
}

type Tag struct {
	Key   *string
	Value *string
	noSmithyDocumentSerde
}

type UnprocessedNamedQueryId struct {
	ErrorCode    *string
	ErrorMessage *string
	NamedQueryId *string
	noSmithyDocumentSerde
}

type UnprocessedQueryExecutionId struct {
	ErrorCode        *string
	ErrorMessage     *string
	QueryExecutionId *string
	noSmithyDocumentSerde
}

type WorkGroup struct {
	Name          *string
	Configuration *WorkGroupConfiguration
	CreationTime  *int64
	Description   *string
	State         types.WorkGroupState
	noSmithyDocumentSerde
}

type WorkGroupConfiguration struct {
	BytesScannedCutoffPerQuery      *int64
	EnforceWorkGroupConfiguration   *bool
	EngineVersion                   *EngineVersion
	PublishCloudWatchMetricsEnabled *bool
	RequesterPaysEnabled            *bool
	ResultConfiguration             *ResultConfiguration
	noSmithyDocumentSerde
}

type WorkGroupConfigurationUpdates struct {
	BytesScannedCutoffPerQuery       *int64
	EnforceWorkGroupConfiguration    *bool
	EngineVersion                    *EngineVersion
	PublishCloudWatchMetricsEnabled  *bool
	RemoveBytesScannedCutoffPerQuery *bool
	RequesterPaysEnabled             *bool
	ResultConfigurationUpdates       *ResultConfigurationUpdates
	noSmithyDocumentSerde
}

type WorkGroupSummary struct {
	CreationTime  *int64
	Description   *string
	EngineVersion *EngineVersion
	Name          *string
	State         types.WorkGroupState
	noSmithyDocumentSerde
}

type noSmithyDocumentSerde = smithydocument.NoSerde

type GetQueryExecutionInput struct {
	QueryExecutionId *string
	noSmithyDocumentSerde
}

type GetQueryExecutionOutput struct {
	QueryExecution *QueryExecution
	ResultMetadata middleware.Metadata
	noSmithyDocumentSerde
}

type StartQueryExecutionInput struct {
	QueryString           *string
	ClientRequestToken    *string
	QueryExecutionContext *QueryExecutionContext
	ResultConfiguration   *ResultConfiguration
	WorkGroup             *string
	noSmithyDocumentSerde
}

type StartQueryExecutionOutput struct {
	QueryExecutionId *string
	ResultMetadata   middleware.Metadata
	noSmithyDocumentSerde
}

type GetQueryResultsInput struct {
	QueryExecutionId *string
	MaxResults       *int32
	NextToken        *string
	noSmithyDocumentSerde
}

type GetQueryResultsOutput struct {
	NextToken      *string
	ResultSet      *ResultSet
	UpdateCount    *int64
	ResultMetadata middleware.Metadata
	noSmithyDocumentSerde
}
