package service

import (
	"errors"

	"github.com/sunmoonstrand/go-react-blog/server/internal/model"
	"gorm.io/gorm"
)

// CreateCategory 创建分类
func CreateCategory(form model.CategoryCreateForm) (int, error) {
	// 检查分类名是否已存在
	var count int64
	if err := model.DB.Model(&model.Category{}).Where("category_name = ?", form.CategoryName).Count(&count).Error; err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("分类名已存在")
	}

	// 检查分类键是否已存在
	if err := model.DB.Model(&model.Category{}).Where("category_key = ?", form.CategoryKey).Count(&count).Error; err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, errors.New("分类键已存在")
	}

	// 创建分类
	category := model.Category{
		CategoryName: form.CategoryName,
		CategoryKey:  form.CategoryKey,
		ParentID:     form.ParentID,
		SortOrder:    form.SortOrder,
		Icon:         form.Icon,
		IsEnabled:    form.IsEnabled,
		Remark:       form.Remark,
	}

	// 保存分类
	if err := model.DB.Create(&category).Error; err != nil {
		return 0, err
	}

	return category.CategoryID, nil
}

// UpdateCategory 更新分类
func UpdateCategory(categoryID int, form model.CategoryUpdateForm) error {
	// 检查分类是否存在
	var category model.Category
	if err := model.DB.First(&category, categoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("分类不存在")
		}
		return err
	}

	// 检查分类名是否已被其他分类使用
	if form.CategoryName != "" && form.CategoryName != category.CategoryName {
		var count int64
		if err := model.DB.Model(&model.Category{}).
			Where("category_name = ? AND category_id != ?", form.CategoryName, categoryID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("分类名已被其他分类使用")
		}
	}

	// 检查分类键是否已被其他分类使用
	if form.CategoryKey != "" && form.CategoryKey != category.CategoryKey {
		var count int64
		if err := model.DB.Model(&model.Category{}).
			Where("category_key = ? AND category_id != ?", form.CategoryKey, categoryID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("分类键已被其他分类使用")
		}
	}

	// 更新分类
	updates := map[string]interface{}{}
	if form.CategoryName != "" {
		updates["category_name"] = form.CategoryName
	}
	if form.CategoryKey != "" {
		updates["category_key"] = form.CategoryKey
	}
	if form.ParentID != nil {
		updates["parent_id"] = *form.ParentID
	}
	if form.SortOrder != nil {
		updates["sort_order"] = *form.SortOrder
	}
	if form.Icon != nil {
		updates["icon"] = *form.Icon
	}
	if form.IsEnabled != nil {
		updates["is_enabled"] = *form.IsEnabled
	}
	if form.Remark != nil {
		updates["remark"] = *form.Remark
	}

	if err := model.DB.Model(&category).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}

// GetCategoryByID 根据ID获取分类
func GetCategoryByID(categoryID int) (*model.Category, error) {
	var category model.Category
	if err := model.DB.First(&category, categoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("分类不存在")
		}
		return nil, err
	}
	return &category, nil
}

// ListCategories 获取分类列表
func ListCategories(params model.CategoryQueryParams) (*model.PageResult, error) {
	var categories []model.Category
	var total int64

	// 构建查询
	query := model.DB.Model(&model.Category{})

	// 应用过滤条件
	if params.CategoryName != "" {
		query = query.Where("category_name LIKE ?", "%"+params.CategoryName+"%")
	}
	if params.CategoryKey != "" {
		query = query.Where("category_key LIKE ?", "%"+params.CategoryKey+"%")
	}
	if params.ParentID != nil {
		query = query.Where("parent_id = ?", *params.ParentID)
	}
	if params.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *params.IsEnabled)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	offset := (params.Page - 1) * params.PageSize
	if err := query.Offset(offset).Limit(params.PageSize).Order("sort_order ASC, category_id ASC").Find(&categories).Error; err != nil {
		return nil, err
	}

	// 转换为响应对象
	var categoryResponses []model.CategoryResponse
	for _, category := range categories {
		categoryResponses = append(categoryResponses, model.CategoryResponse{
			CategoryID:   category.CategoryID,
			CategoryName: category.CategoryName,
			CategoryKey:  category.CategoryKey,
			ParentID:     category.ParentID,
			SortOrder:    category.SortOrder,
			Icon:         category.Icon,
			IsEnabled:    category.IsEnabled,
			Remark:       category.Remark,
			CreatedAt:    category.CreatedAt,
			UpdatedAt:    category.UpdatedAt,
		})
	}

	return model.NewPageResult(categoryResponses, total, params.Page, params.PageSize), nil
}

// DeleteCategory 删除分类
func DeleteCategory(categoryID int) error {
	// 检查分类是否存在
	var category model.Category
	if err := model.DB.First(&category, categoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("分类不存在")
		}
		return err
	}

	// 检查是否有子分类
	var count int64
	if err := model.DB.Model(&model.Category{}).Where("parent_id = ?", categoryID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该分类下有子分类，不能删除")
	}

	// 检查是否有关联的文章
	if err := model.DB.Model(&model.Article{}).Where("category_id = ?", categoryID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该分类下有文章，不能删除")
	}

	// 删除分类
	if err := model.DB.Delete(&category).Error; err != nil {
		return err
	}

	return nil
}

// GetAllCategories 获取所有分类（用于下拉选择）
func GetAllCategories() ([]model.Option, error) {
	var categories []model.Category
	if err := model.DB.Where("is_enabled = ?", true).Order("sort_order ASC, category_id ASC").Find(&categories).Error; err != nil {
		return nil, err
	}

	var options []model.Option
	for _, category := range categories {
		options = append(options, model.Option{
			Label: category.CategoryName,
			Value: category.CategoryID,
		})
	}

	return options, nil
}

// GetCategoryTree 获取分类树
func GetCategoryTree() ([]*model.TreeNode, error) {
	var categories []model.Category
	if err := model.DB.Order("sort_order ASC, category_id ASC").Find(&categories).Error; err != nil {
		return nil, err
	}

	// 构建树结构
	categoryMap := make(map[int]*model.TreeNode)
	for _, category := range categories {
		categoryMap[category.CategoryID] = &model.TreeNode{
			ID:       category.CategoryID,
			Label:    category.CategoryName,
			Value:    category.CategoryID,
			Children: []*model.TreeNode{},
		}
	}

	var rootNodes []*model.TreeNode
	for _, category := range categories {
		node := categoryMap[category.CategoryID]
		if category.ParentID == nil || *category.ParentID == 0 {
			// 根节点
			rootNodes = append(rootNodes, node)
		} else {
			// 子节点
			if parent, ok := categoryMap[*category.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return rootNodes, nil
}

// UpdateCategoryStatus 更新分类状态
func UpdateCategoryStatus(categoryID int, isEnabled bool) error {
	// 检查分类是否存在
	var category model.Category
	if err := model.DB.First(&category, categoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("分类不存在")
		}
		return err
	}

	// 更新状态
	if err := model.DB.Model(&category).Update("is_enabled", isEnabled).Error; err != nil {
		return err
	}

	return nil
}
