package usecase

import (
	"Brewery/internal/apperrors"
	"Brewery/internal/entities"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const professionHourlyRateField = "hour_salary"

// OperationRepository определяет доступ к хранилищу технологических операций.
// В результатах InsertOperation, GetOperations, GetOperationByID и UpdateOperation
// хранилище заполняет идентификаторы материала и изделия, а расход и Base заполняет сервис.
type OperationRepository interface {
	// InsertOperation сохраняет операцию и ее параметры, используя существующую связь в составе изделия.
	InsertOperation(ctx context.Context, op *entities.Operation, values []entities.ParameterValue) (*entities.Operation, error)

	// GetOperations возвращает операции с пагинацией и необязательными фильтрами по материалу и продукту.
	GetOperations(ctx context.Context, limit, offset int32, materialID, productID *int64) ([]entities.Operation, error)

	// GetOperationByID возвращает операцию с оборудованием, параметрами и идентификаторами материала и изделия.
	GetOperationByID(ctx context.Context, id int64) (*entities.Operation, error)

	// UpdateOperation применяет изменения атомарно и возвращает сохраненную операцию.
	// Поля equipment и materials передаются вложенными map[string]any с указанными подполями.
	UpdateOperation(ctx context.Context, id int64, updates map[string]any, changes entities.ParameterChanges) (*entities.Operation, error)

	// DeleteOperation удаляет операцию по ID. Если запись отсутствует, возвращает ошибку.
	DeleteOperation(ctx context.Context, id int64) error

	// GetProductOperations возвращает операции, у которых указанный продукт является выходным.
	GetProductOperations(ctx context.Context, productID int64) ([]entities.Operation, error)
}

// OperationService реализует бизнес-логику управления технологическими операциями.
type OperationService struct {
	Repo                OperationRepository
	MaterialRepo        MaterialRepository
	ProductRepo         ProductRepository
	ClassRepo           ClassRepository
	ClassParametersRepo ClassParametersRepository
	EnumValuesRepo      EnumValuesRepository
}

// NewOperationService creates operation service.
func NewOperationService(repo OperationRepository, materialRepo MaterialRepository,
	productsRepo ProductRepository, classRepo ClassRepository,
	classParamsRepo ClassParametersRepository, enumValuesRepo EnumValuesRepository) *OperationService {
	return &OperationService{
		Repo:                repo,
		ProductRepo:         productsRepo,
		MaterialRepo:        materialRepo,
		ClassRepo:           classRepo,
		ClassParametersRepo: classParamsRepo,
		EnumValuesRepo:      enumValuesRepo,
	}
}

// CreateOperation проверяет данные и сохраняет операцию.
func (s *OperationService) CreateOperation(ctx context.Context, op entities.Operation) (*entities.Operation, error) {
	switch {
	case strings.TrimSpace(op.Name) == "":
		return nil, apperrors.BadRequest("name is required", nil)
	case op.ClassID <= 0:
		return nil, apperrors.BadRequest("operation_class_id must be positive", nil)
	case op.ProfessionID <= 0:
		return nil, apperrors.BadRequest("profession_id must be positive", nil)
	case strings.TrimSpace(op.Equipment.Name) == "" || strings.TrimSpace(op.Equipment.WorkCenter) == "":
		return nil, apperrors.BadRequest("equipment.name and equipment.work_center are required", nil)
	case op.StepNumber <= 0:
		return nil, apperrors.BadRequest("step_number must be positive", nil)
	case !validOperationDecimal(op.LaborCost):
		return nil, apperrors.BadRequest("invalid labor_cost", nil)
	case op.Materials.MaterialID <= 0 || op.Materials.ProductID <= 0:
		return nil, apperrors.BadRequest("materials.material_id and materials.product_id must be positive", nil)
	case op.Materials.MaterialID == op.Materials.ProductID:
		return nil, apperrors.BadRequest("input and output products must be different", nil)
	case !validOperationDecimal(float64(op.Materials.MaterialAmount)):
		return nil, apperrors.BadRequest("invalid materials.material_amount", nil)
	}
	if err := s.validateOperationClasses(ctx, op.ClassID, op.ProfessionID); err != nil {
		return nil, err
	}
	values, canonical, err := resolveClassParameters(ctx, s.ClassParametersRepo, s.EnumValuesRepo, op.ClassID, op.Parameters)
	if err != nil {
		return nil, err
	}
	op.Parameters = canonical

	materials, err := s.loadOperationMaterials(ctx, op.Materials.ProductID)
	if err != nil {
		return nil, err
	}
	material, ok := materials[op.Materials.MaterialID]
	if !ok {
		return nil, apperrors.BadRequest("material is not part of the output product", nil)
	}

	created, err := s.Repo.InsertOperation(ctx, &op, values)
	if err != nil {
		return nil, operationWriteError(err)
	}
	if created == nil {
		return nil, errors.New("operation repository returned no operation after creation")
	}
	created.Materials.MaterialAmount = material.MaterialAmount
	created.Materials.Base = material.Base
	return created, nil
}

// GetOperations возвращает страницу операций с фильтрами по входному и выходному продукту.
func (s *OperationService) GetOperations(ctx context.Context, limit, offset int32, materialID, productID *int64) ([]entities.Operation, error) {
	if limit <= 0 || offset < 0 {
		return nil, apperrors.BadRequest("invalid pagination", nil)
	}
	if materialID != nil && *materialID <= 0 {
		return nil, apperrors.BadRequest("invalid material_id", nil)
	}
	if productID != nil && *productID <= 0 {
		return nil, apperrors.BadRequest("invalid product_id", nil)
	}

	ops, err := s.Repo.GetOperations(ctx, limit, offset, materialID, productID)
	if err != nil {
		return nil, err
	}
	materialsByProduct := make(map[int64]map[int64]entities.Material)
	for i := range ops {
		if err := s.fillOperationMaterial(ctx, &ops[i], materialsByProduct); err != nil {
			return nil, err
		}
	}
	return ops, nil
}

// GetOperationByID возвращает операцию и связанные данные.
func (s *OperationService) GetOperationByID(ctx context.Context, id int64) (*entities.Operation, error) {
	if id <= 0 {
		return nil, apperrors.BadRequest("invalid operation id", nil)
	}

	op, err := s.getOperation(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.fillOperationMaterial(ctx, op, nil); err != nil {
		return nil, err
	}
	return op, nil
}

// DeleteOperation удаляет операцию по ID.
func (s *OperationService) DeleteOperation(ctx context.Context, id int64) error {
	if id <= 0 {
		return apperrors.BadRequest("invalid operation id", nil)
	}

	return operationLookupError(s.Repo.DeleteOperation(ctx, id))
}

func (s *OperationService) getOperation(ctx context.Context, id int64) (*entities.Operation, error) {
	op, err := s.Repo.GetOperationByID(ctx, id)
	if err != nil {
		return nil, operationLookupError(err)
	}
	if op == nil {
		return nil, apperrors.NotFound("operation not found", nil)
	}
	return op, nil
}

func (s *OperationService) validateOperationClasses(ctx context.Context, classID, professionID int64) error {
	if s.ClassRepo == nil {
		return errors.New("class repository is nil")
	}
	if err := s.validateOperationClassBranch(ctx, classID, "operation_class_id", "operation"); err != nil {
		return err
	}
	return s.validateOperationClassBranch(ctx, professionID, "profession_id", "profession")
}

func (s *OperationService) validateOperationClassBranch(ctx context.Context, id int64, field, branch string) error {
	visited := make(map[int64]struct{})
	for currentID := id; currentID > 0; {
		if _, seen := visited[currentID]; seen {
			return fmt.Errorf("class hierarchy contains a cycle at class %d", currentID)
		}
		visited[currentID] = struct{}{}

		class, err := s.ClassRepo.GetClassByID(ctx, currentID)
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && class == nil) {
			return apperrors.BadRequest(fmt.Sprintf("%s refers to a missing class", field), err)
		}
		if err != nil {
			return fmt.Errorf("get class %d while validating %s: %w", currentID, field, err)
		}
		if class.Name == branch {
			return nil
		}
		if class.ParentID == nil {
			break
		}
		currentID = *class.ParentID
	}
	return apperrors.BadRequest(fmt.Sprintf("%s must belong to the %s branch", field, branch), nil)
}

// UpdateOperation проверяет и применяет частичные изменения операции.
func (s *OperationService) UpdateOperation(ctx context.Context, id int64, updates map[string]any) (*entities.Operation, error) {
	if id <= 0 {
		return nil, apperrors.BadRequest("invalid operation id", nil)
	}

	normalized, err := normalizeOperationUpdates(updates)
	if err != nil {
		return nil, err
	}
	current, err := s.getOperation(ctx, id)
	if err != nil {
		return nil, err
	}
	classID, professionID := current.ClassID, current.ProfessionID
	if updatedClassID, ok := normalized["operation_class_id"].(int64); ok {
		classID = updatedClassID
	}
	if updatedProfessionID, ok := normalized["profession_id"].(int64); ok {
		professionID = updatedProfessionID
	}
	if err := s.validateOperationClasses(ctx, classID, professionID); err != nil {
		return nil, err
	}

	patch, _ := normalized["parameters"].(map[string]any)
	delete(normalized, "parameters")
	changes, candidate := prepareParameterPatch(current.ClassID, classID, current.Parameters, patch)
	changes.Values, _, err = resolveClassParameters(ctx, s.ClassParametersRepo, s.EnumValuesRepo, classID, candidate)
	if err != nil {
		return nil, err
	}

	op, err := s.Repo.UpdateOperation(ctx, id, normalized, changes)
	if err != nil {
		return nil, operationWriteError(err)
	}
	if op == nil {
		return nil, apperrors.NotFound("operation not found", nil)
	}
	if err := s.fillOperationMaterial(ctx, op, nil); err != nil {
		return nil, fmt.Errorf("operation %d updated but material data could not be loaded: %w", id, err)
	}
	return op, nil
}

func (s *OperationService) loadOperationMaterials(ctx context.Context, productID int64) (map[int64]entities.Material, error) {
	rows, err := s.MaterialRepo.GetMaterialsByProduct(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("get materials for product %d: %w", productID, err)
	}
	materials := make(map[int64]entities.Material, len(rows))
	for _, row := range rows {
		materials[row.MaterialID] = row.Material
	}
	return materials, nil
}

func (s *OperationService) fillOperationMaterial(ctx context.Context, op *entities.Operation, materialsByProduct map[int64]map[int64]entities.Material) error {
	productID := op.Materials.ProductID
	materials, loaded := materialsByProduct[productID]
	if !loaded {
		var err error
		materials, err = s.loadOperationMaterials(ctx, productID)
		if err != nil {
			return err
		}
		if materialsByProduct != nil {
			materialsByProduct[productID] = materials
		}
	}
	material, ok := materials[op.Materials.MaterialID]
	if !ok {
		return fmt.Errorf("material %d for operation %d is missing from product %d", op.Materials.MaterialID, op.ID, productID)
	}
	op.Materials.MaterialAmount = material.MaterialAmount
	op.Materials.Base = material.Base
	return nil
}

// GetProductTechCard собирает операции, нормы базовых материалов и стоимость производства.
func (s *OperationService) GetProductTechCard(ctx context.Context, productID int64, quantity float32) (*entities.ProductTechCard, error) {
	if productID <= 0 {
		return nil, apperrors.BadRequest("invalid product id", nil)
	}
	if quantity <= 0 || math.IsNaN(float64(quantity)) || math.IsInf(float64(quantity), 0) {
		return nil, apperrors.BadRequest("quantity must be a finite positive number", nil)
	}

	product, err := s.ProductRepo.GetProductByID(ctx, productID)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && product == nil) {
		return nil, apperrors.NotFound("product not found", err)
	}
	if err != nil {
		return nil, fmt.Errorf("get product for tech card: %w", err)
	}

	materials := NewMaterialService(s.MaterialRepo, s.ProductRepo)
	tree, err := materials.GetMaterialsByProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	operations, requiredQuantities, err := s.collectTechCardOperations(ctx, productID, float64(quantity), tree)
	if err != nil {
		return nil, err
	}
	norms, err := collectMaterialNorms(tree, quantity, techCardBaseMaterials)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(norms, func(a, b entities.MaterialProduct) int { return cmp.Compare(a.MaterialID, b.MaterialID) })
	var materialCost float64
	for _, norm := range norms {
		if norm.IsPurchased {
			materialCost += float64(norm.MaterialAmount) * norm.Price
		}
	}

	rates := make(map[int64]float64)
	var laborCost float64
	for _, operation := range operations {
		rate, ok := rates[operation.ProfessionID]
		if !ok {
			rate, err = s.professionHourlyRate(ctx, operation.ProfessionID)
			if err != nil {
				return nil, err
			}
			rates[operation.ProfessionID] = rate
		}
		laborCost += rate * operation.LaborCost * requiredQuantities[operation.Materials.ProductID]
	}
	total := materialCost + laborCost
	if math.IsNaN(total) || math.IsInf(total, 0) {
		return nil, apperrors.Validation("calculated production cost is out of range", nil)
	}
	return &entities.ProductTechCard{
		ProductID:     productID,
		TotalCost:     total,
		Operations:    operations,
		MaterialNorms: norms,
	}, nil
}

func (s *OperationService) professionHourlyRate(ctx context.Context, professionID int64) (float64, error) {
	if professionID <= 0 {
		return 0, apperrors.Validation("invalid profession id in operation", nil)
	}
	if s.ClassParametersRepo == nil {
		return 0, errors.New("class parameters repository is nil")
	}
	parameters, _, err := s.ClassParametersRepo.GetParameters(ctx, professionID, entities.RangeParameterType)
	if err != nil {
		return 0, fmt.Errorf("get range parameters for profession %d: %w", professionID, err)
	}

	var rate float64
	found := false
	for _, parameter := range parameters {
		if parameter.FieldName != professionHourlyRateField {
			continue
		}
		if found || parameter.MinValue < 0 {
			return 0, apperrors.Validation(fmt.Sprintf("invalid %s range parameter for profession %d", professionHourlyRateField, professionID), nil)
		}
		rate = float64(parameter.MinValue)
		found = true
	}
	if !found {
		return 0, apperrors.Validation(fmt.Sprintf("profession %d is missing the %s range parameter", professionID, professionHourlyRateField), nil)
	}
	return rate, nil
}

func (s *OperationService) collectTechCardOperations(ctx context.Context, productID int64, quantity float64, tree []entities.MaterialProduct) ([]entities.Operation, map[int64]float64, error) {
	result := make([]entities.Operation, 0)
	materialsByProduct := make(map[int64][]entities.MaterialProduct)
	productionOrder := make([]int64, 0)
	var collect func(id int64, materials []entities.MaterialProduct) error
	collect = func(id int64, materials []entities.MaterialProduct) error {
		if _, visited := materialsByProduct[id]; visited {
			return nil
		}
		materialsByProduct[id] = materials

		slices.SortFunc(materials, func(a, b entities.MaterialProduct) int { return cmp.Compare(a.MaterialID, b.MaterialID) })
		materialsByID := make(map[int64]entities.Material, len(materials))
		for _, material := range materials {
			materialID := material.MaterialID
			if materialID == 0 {
				materialID = material.ID
			}
			materialsByID[materialID] = material.Material
			// Без входных материалов у изделия нет операций производства.
			if len(material.Materials) == 0 {
				continue
			}
			if err := collect(materialID, material.Materials); err != nil {
				return err
			}
		}

		operations, err := s.Repo.GetProductOperations(ctx, id)
		if err != nil {
			return fmt.Errorf("get operations for product %d: %w", id, err)
		}
		operations = slices.Clone(operations)
		slices.SortFunc(operations, func(a, b entities.Operation) int { return cmp.Compare(a.StepNumber, b.StepNumber) })
		for i := range operations {
			operation := &operations[i]
			material, ok := materialsByID[operation.Materials.MaterialID]
			if !ok {
				return fmt.Errorf("material %d for operation %d is missing from product %d", operation.Materials.MaterialID, operation.ID, id)
			}
			operation.Materials.MaterialAmount = material.MaterialAmount
			operation.Materials.Base = material.Base
		}
		result = append(result, operations...)
		productionOrder = append(productionOrder, id)
		return nil
	}
	if err := collect(productID, tree); err != nil {
		return nil, nil, err
	}

	// Обратный порядок учитывает потребность всех родителей до расчёта количества полуфабриката.
	required := map[int64]float64{productID: quantity}
	for _, id := range slices.Backward(productionOrder) {
		productQuantity := required[id]
		if math.IsNaN(productQuantity) || math.IsInf(productQuantity, 0) {
			return nil, nil, apperrors.Validation("calculated production quantity is out of range", nil)
		}
		for _, material := range materialsByProduct[id] {
			if len(material.Materials) == 0 {
				continue
			}
			materialID := material.MaterialID
			if materialID == 0 {
				materialID = material.ID
			}
			required[materialID] += productQuantity * float64(material.MaterialAmount)
		}
	}
	return result, required, nil
}

func normalizeOperationUpdates(updates map[string]any) (map[string]any, error) {
	if len(updates) == 0 {
		return nil, apperrors.BadRequest("empty operation update", nil)
	}

	normalized := make(map[string]any, len(updates))
	for field, value := range updates {
		if strings.Contains(field, ".") {
			return nil, apperrors.BadRequest(fmt.Sprintf("unknown operation update field %q", field), nil)
		}
		var err error
		normalized[field], err = normalizeOperationUpdate(field, value)
		if err != nil {
			return nil, err
		}
	}
	return normalized, nil
}

func normalizeOperationUpdate(field string, value any) (any, error) {
	switch field {
	case "name":
		if name, ok := value.(string); ok && strings.TrimSpace(name) != "" {
			return name, nil
		}
	case "description":
		if description, ok := value.(string); ok {
			return description, nil
		}
	case "operation_class_id", "profession_id":
		if id, ok := operationPositiveInt(value); ok {
			return id, nil
		}
	case "step_number":
		if step, ok := operationPositiveInt(value); ok && step > 0 && step <= math.MaxInt32 {
			return int32(step), nil
		}
	case "labor_cost":
		if cost, ok := operationNumber(value); ok && validOperationDecimal(cost) {
			return cost, nil
		}
	case "parameters":
		if parameters, ok := value.(map[string]any); ok && parameters != nil {
			return parameters, nil
		}
	case "equipment", "materials":
		if fields, ok := value.(map[string]any); ok && len(fields) > 0 {
			nested := make(map[string]any, len(fields))
			for key, val := range fields {
				path := field + "." + key
				var err error
				nested[key], err = normalizeOperationUpdate(path, val)
				if err != nil {
					return nil, err
				}
			}
			return nested, nil
		}
	case "equipment.name", "equipment.work_center":
		if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
			return text, nil
		}
	case "materials.material_id", "materials.product_id":
		if id, ok := operationPositiveInt(value); ok {
			return id, nil
		}
	default:
		return nil, apperrors.BadRequest(fmt.Sprintf("unknown operation update field %q", field), nil)
	}
	return nil, apperrors.BadRequest(fmt.Sprintf("invalid operation update field %q", field), nil)
}

func operationPositiveInt(value any) (int64, bool) {
	switch number := value.(type) {
	case json.Number:
		id, err := number.Int64()
		return id, err == nil && id > 0
	case float64:
		if number < 1 || number >= 1<<63 || number != math.Trunc(number) {
			return 0, false
		}
	case float32:
		return operationPositiveInt(float64(number))
	}
	id, ok := toInt64Value(value)
	return id, ok && id > 0
}

func operationNumber(value any) (float64, bool) {
	if number, ok := value.(json.Number); ok {
		parsed, err := number.Float64()
		return parsed, err == nil && !math.IsInf(parsed, 0) && !math.IsNaN(parsed)
	}
	number, ok := toFloat64Value(value)
	return number, ok && !math.IsInf(number, 0) && !math.IsNaN(number)
}

func validOperationDecimal(value float64) bool {
	return value >= 0 && value <= 9_999_999.999
}

func operationLookupError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperrors.NotFound("operation not found", err)
	}
	return err
}

func operationWriteError(err error) error {
	if _, ok := errors.AsType[*apperrors.AppError](err); ok {
		return err
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.ConstraintName {
		case "idx_operations_output_product_step":
			return apperrors.Conflict("step_number is already used for this product", err)
		case "fk_operations_work_center":
			return apperrors.BadRequest("equipment with this name and work_center does not exist", err)
		case "fk_operations_material":
			return apperrors.BadRequest("material is not part of the output product", err)
		case "operations_operation_class_id_fkey":
			return apperrors.BadRequest("operation class does not exist", err)
		case "operations_profession_id_fkey":
			return apperrors.BadRequest("profession does not exist", err)
		case "operations_input_product_id_fkey", "operations_output_product_id_fkey":
			return apperrors.BadRequest("input or output product does not exist", err)
		}
	}

	return operationLookupError(err)
}
