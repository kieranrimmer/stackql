package drm

import (
	"strings"

	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
	"github.com/stackql/stackql/internal/stackql/sql_system"
	"github.com/stackql/stackql/internal/stackql/tablenamespace"
)

type PreparedStatementCtx interface {
	GetAllCtrlCtrs() []internaldto.TxnControlCounters
	GetGCCtrlCtrs() internaldto.TxnControlCounters
	GetIndirectContexts() []PreparedStatementCtx
	GetNonControlColumns() []internaldto.ColumnMetadata
	GetGCHousekeepingQueries() string
	GetQuery() string
	SetGCCtrlCtrs(tcc internaldto.TxnControlCounters)
	SetIndirectContexts(indirectContexts []PreparedStatementCtx)
	SetKind(kind string)
}

type standardPreparedStatementCtx struct {
	query                   string
	kind                    string // string annotation applicable only in some cases eg UNION [ALL]
	genIdControlColName     string
	sessionIdControlColName string
	TableNames              []string
	txnIdControlColName     string
	insIdControlColName     string
	insEncodedColName       string
	nonControlColumns       []internaldto.ColumnMetadata
	ctrlColumnRepeats       int
	txnCtrlCtrs             internaldto.TxnControlCounters
	selectTxnCtrlCtrs       []internaldto.TxnControlCounters
	namespaceCollection     tablenamespace.TableNamespaceCollection
	sqlSystem               sql_system.SQLSystem
	indirectContexts        []PreparedStatementCtx
}

func (ps *standardPreparedStatementCtx) SetIndirectContexts(indirectContexts []PreparedStatementCtx) {
	ps.indirectContexts = indirectContexts
}

func (ps *standardPreparedStatementCtx) GetIndirectContexts() []PreparedStatementCtx {
	return ps.indirectContexts
}

func (ps *standardPreparedStatementCtx) SetKind(kind string) {
	ps.kind = kind
}

func (ps *standardPreparedStatementCtx) GetQuery() string {
	return ps.query
}

func (ps *standardPreparedStatementCtx) GetGCCtrlCtrs() internaldto.TxnControlCounters {
	return ps.txnCtrlCtrs
}

func (ps *standardPreparedStatementCtx) SetGCCtrlCtrs(tcc internaldto.TxnControlCounters) {
	ps.txnCtrlCtrs = tcc
}

func (ps *standardPreparedStatementCtx) GetNonControlColumns() []internaldto.ColumnMetadata {
	return ps.nonControlColumns
}

func (ps *standardPreparedStatementCtx) GetAllCtrlCtrs() []internaldto.TxnControlCounters {
	var rv []internaldto.TxnControlCounters
	rv = append(rv, ps.txnCtrlCtrs)
	rv = append(rv, ps.selectTxnCtrlCtrs...)
	return rv
}

func NewPreparedStatementCtx(
	query string,
	kind string,
	genIdControlColName string,
	sessionIdControlColName string,
	tableNames []string,
	txnIdControlColName string,
	insIdControlColName string,
	insEncodedColName string,
	nonControlColumns []internaldto.ColumnMetadata,
	ctrlColumnRepeats int,
	txnCtrlCtrs internaldto.TxnControlCounters,
	secondaryCtrs []internaldto.TxnControlCounters,
	namespaceCollection tablenamespace.TableNamespaceCollection,
	sqlSystem sql_system.SQLSystem,
) PreparedStatementCtx {
	return &standardPreparedStatementCtx{
		query:                   query,
		kind:                    kind,
		genIdControlColName:     genIdControlColName,
		sessionIdControlColName: sessionIdControlColName,
		TableNames:              tableNames,
		txnIdControlColName:     txnIdControlColName,
		insIdControlColName:     insIdControlColName,
		insEncodedColName:       insEncodedColName,
		nonControlColumns:       nonControlColumns,
		ctrlColumnRepeats:       ctrlColumnRepeats,
		txnCtrlCtrs:             txnCtrlCtrs,
		selectTxnCtrlCtrs:       secondaryCtrs,
		namespaceCollection:     namespaceCollection,
		sqlSystem:               sqlSystem,
	}
}

func NewQueryOnlyPreparedStatementCtx(query string, nonControlCols []internaldto.ColumnMetadata) PreparedStatementCtx {
	return &standardPreparedStatementCtx{query: query, nonControlColumns: nonControlCols, ctrlColumnRepeats: 0}
}

func (ps *standardPreparedStatementCtx) GetGCHousekeepingQueries() string {
	var housekeepingQueries []string
	for _, table := range ps.TableNames {
		housekeepingQueries = append(housekeepingQueries, ps.sqlSystem.GetGCHousekeepingQuery(table, ps.txnCtrlCtrs))
	}
	return strings.Join(housekeepingQueries, "; ")
}
