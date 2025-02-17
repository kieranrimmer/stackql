package dataflow

import (
	"github.com/stackql/go-openapistackql/openapistackql"
	"github.com/stackql/stackql-parser/go/vt/sqlparser"
	"gonum.org/v1/gonum/graph"
)

type DataFlowEdge interface {
	graph.WeightedEdge
	AddRelation(DataFlowRelation)
	GetColumnDescriptors() ([]openapistackql.ColumnDescriptor, error)
	GetDest() DataFlowVertex
	GetProjection() (map[string]string, error)
	GetSelectExprs() (sqlparser.SelectExprs, error)
	GetSource() DataFlowVertex
	IsSQL() bool
}

type standardDataFlowEdge struct {
	source, dest DataFlowVertex
	relations    []DataFlowRelation
}

func NewStandardDataFlowEdge(
	source DataFlowVertex,
	dest DataFlowVertex,
	comparisonExpr *sqlparser.ComparisonExpr,
	sourceExpr sqlparser.Expr,
	destColumn *sqlparser.ColName,
) DataFlowEdge {
	return &standardDataFlowEdge{
		source: source,
		dest:   dest,
		relations: []DataFlowRelation{
			NewStandardDataFlowRelation(
				comparisonExpr,
				destColumn,
				sourceExpr,
			),
		},
	}
}

func (de *standardDataFlowEdge) AddRelation(rel DataFlowRelation) {
	de.relations = append(de.relations, rel)
}

func (de *standardDataFlowEdge) From() graph.Node {
	return de.source
}

func (de *standardDataFlowEdge) To() graph.Node {
	return de.dest
}

func (de *standardDataFlowEdge) ReversedEdge() graph.Edge {
	// Reversal is invalid given the assymetric
	// expressions, therefore returning unaltered
	// as per library recommmendation.
	return de
}

func (de *standardDataFlowEdge) Weight() float64 {
	return 1.0
}

func (de *standardDataFlowEdge) GetSource() DataFlowVertex {
	return de.source
}

func (de *standardDataFlowEdge) IsSQL() bool {
	for _, rel := range de.relations {
		if rel.IsSQL() {
			return true
		}
	}
	return false
}

func (de *standardDataFlowEdge) GetDest() DataFlowVertex {
	return de.dest
}

func (dv *standardDataFlowEdge) GetProjection() (map[string]string, error) {
	rv := make(map[string]string)
	for _, rel := range dv.relations {
		src, dst, err := rel.GetProjection()
		if err != nil {
			return nil, err
		}
		rv[src] = dst
	}
	return rv, nil
}

func (dv *standardDataFlowEdge) GetSelectExprs() (sqlparser.SelectExprs, error) {
	var rv sqlparser.SelectExprs
	for _, rel := range dv.relations {
		selExpr, err := rel.GetSelectExpr()
		if err != nil {
			return nil, err
		}
		rv = append(rv, selExpr)
	}
	return rv, nil
}

func (dv *standardDataFlowEdge) GetColumnDescriptors() ([]openapistackql.ColumnDescriptor, error) {
	var rv []openapistackql.ColumnDescriptor
	for _, rel := range dv.relations {
		d, err := rel.GetColumnDescriptor()
		if err != nil {
			return nil, err
		}
		rv = append(rv, d)
	}
	return rv, nil
}
