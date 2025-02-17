package taxonomy

import (
	"github.com/stackql/go-openapistackql/openapistackql"
	"github.com/stackql/stackql/internal/stackql/handler"
	"github.com/stackql/stackql/internal/stackql/httpbuild"
	"github.com/stackql/stackql/internal/stackql/internal_data_transfer/internaldto"
	"github.com/stackql/stackql/internal/stackql/logging"
	"github.com/stackql/stackql/internal/stackql/streaming"
	"github.com/stackql/stackql/internal/stackql/tablemetadata"
	"github.com/stackql/stackql/internal/stackql/util"
)

// TODO:
//   - For views, need API to get child.
type AnnotationCtx interface {
	GetHIDs() internaldto.HeirarchyIdentifiers
	IsDynamic() bool
	GetView() (internaldto.ViewDTO, bool)
	GetSubquery() (internaldto.SubqueryDTO, bool)
	GetInputTableName() (string, error)
	GetParameters() map[string]interface{}
	GetSchema() *openapistackql.Schema
	GetTableMeta() tablemetadata.ExtendedTableMetadata
	Prepare(handlerCtx handler.HandlerContext, inStream streaming.MapStream) error
	SetDynamic()
}

type standardAnnotationCtx struct {
	isDynamic  bool
	schema     *openapistackql.Schema
	hIDs       internaldto.HeirarchyIdentifiers
	tableMeta  tablemetadata.ExtendedTableMetadata
	parameters map[string]interface{}
}

func NewStaticStandardAnnotationCtx(
	schema *openapistackql.Schema,
	hIds internaldto.HeirarchyIdentifiers,
	tableMeta tablemetadata.ExtendedTableMetadata,
	parameters map[string]interface{},
) AnnotationCtx {
	return &standardAnnotationCtx{
		isDynamic:  false,
		schema:     schema,
		hIDs:       hIds,
		tableMeta:  tableMeta,
		parameters: parameters,
	}
}

func (ac *standardAnnotationCtx) IsDynamic() bool {
	return ac.isDynamic
}

func (ac *standardAnnotationCtx) GetView() (internaldto.ViewDTO, bool) {
	return ac.hIDs.GetView()
}

func (ac *standardAnnotationCtx) GetSubquery() (internaldto.SubqueryDTO, bool) {
	return ac.hIDs.GetSubquery()
}

func (ac *standardAnnotationCtx) SetDynamic() {
	ac.isDynamic = true
}

func (ac *standardAnnotationCtx) Prepare(
	handlerCtx handler.HandlerContext,
	stream streaming.MapStream,
) error {
	// TODO: accomodate SQL data source
	sqlDataSource, isSQLDataSource := ac.GetTableMeta().GetSQLDataSource()
	if isSQLDataSource {
		ac.tableMeta.SetSQLDataSource(sqlDataSource)
		// TODO: persist mirror table here a la GenerateInsertDML()
		// anTab := util.NewAnnotatedTabulation(tab, ac.GetHIDs(), inputTableName, annotationCtx.GetTableMeta().GetAlias())
		// ddl, err := handlerCtx.GetDrmConfig().GenerateDDL(ac.tableMeta, nil, 0, false)
		return nil
	}
	pr, err := ac.GetTableMeta().GetProvider()
	if err != nil {
		return err
	}
	svc, err := ac.GetTableMeta().GetService()
	if err != nil {
		return err
	}
	opStore, err := ac.GetTableMeta().GetMethod()
	if err != nil {
		return err
	}
	// LAZY EVAL if dynamic
	if ac.isDynamic {
		viewDTO, isView := ac.GetView()
		// TODO: fill this out
		if isView {
			logging.GetLogger().Debugf("viewDTO = %v\n", viewDTO)
		}
		ac.tableMeta.WithGetHttpArmoury(
			func() (httpbuild.HTTPArmoury, error) {
				httpArmoury, err := httpbuild.BuildHTTPRequestCtxFromAnnotation(stream, pr, opStore, svc, nil, nil)
				return httpArmoury, err
			},
		)
		return nil
	} else {
		// moved out of here so stream is dynamically generated
	}
	ac.tableMeta.WithGetHttpArmoury(
		func() (httpbuild.HTTPArmoury, error) {
			// need to dynamically generate stream, otherwise repeated calls result in empty body
			parametersCleaned, err := util.TransformSQLRawParameters(ac.GetParameters())
			if err != nil {
				return nil, err
			}
			stream.Write(
				[]map[string]interface{}{
					parametersCleaned,
				},
			)
			httpArmoury, err := httpbuild.BuildHTTPRequestCtxFromAnnotation(stream, pr, opStore, svc, nil, nil)
			if err != nil {
				return nil, err
			}
			return httpArmoury, nil
		},
	)
	return nil
}

func (ac *standardAnnotationCtx) GetHIDs() internaldto.HeirarchyIdentifiers {
	return ac.hIDs
}

func (ac *standardAnnotationCtx) GetParameters() map[string]interface{} {
	return ac.parameters
}

func (ac *standardAnnotationCtx) GetSchema() *openapistackql.Schema {
	return ac.schema
}

func (ac *standardAnnotationCtx) GetInputTableName() (string, error) {
	return ac.tableMeta.GetInputTableName()
}

func (ac *standardAnnotationCtx) GetTableMeta() tablemetadata.ExtendedTableMetadata {
	return ac.tableMeta
}
