package main

import (
	"database/sql"
)

type Category struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Products []*Product `json:"products"`
}

type Product struct {
	ID               string    `json:"id"`
	Category         *Category `json:"category"`
	SubCategoryID    string    `json:"sub_category_id"`
	SubCategoryName  string    `json:"sub_category_name"`
	Name             string    `json:"name"`
	ShortDescription string    `json:"short_description"`
	Description      string    `json:"description"`
	MainPurpose      string    `json:"main_purpose"`
	Features         string    `json:"features"`
	FormulaRef       string    `json:"formula_ref"`
	ERPID            string    `json:"erp_id"`
}

type ProductsRepository struct {
	Repository
	db *sql.DB
}

func NewProductsRepository(db *sql.DB) *ProductsRepository {
	return &ProductsRepository{
		db: db,
	}
}

func NewCategory(id, name string) *Category {
	return &Category{
		ID:       id,
		Name:     name,
		Products: make([]*Product, 0),
	}
}

func NewProduct(id string, category *Category, subCategoryID, subCategoryName, name, shortDescription, description, mainPurpose, features, formulaRef string, erpID string) *Product {
	return &Product{
		ID:               id,
		Category:         category,
		SubCategoryID:    subCategoryID,
		SubCategoryName:  subCategoryName,
		Name:             name,
		ShortDescription: shortDescription,
		Description:      description,
		MainPurpose:      mainPurpose,
		Features:         features,
		FormulaRef:       formulaRef,
		ERPID:            erpID,
	}
}

func (p *ProductsRepository) GetProducts() ([]*Product, error) {
	var products []*Product
	rows, err := p.db.Query("exec Mobile_GetProductsFlat;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id               string
			categoryID       string
			cateogryName     string
			subCategoryID    string
			subCategoryName  string
			name             string
			shortDescription string
			description      string
			mainPurpose      string
			features         string
			formulaRef       string
			erpID            string
		)
		rows.Scan(&id, &categoryID, &cateogryName, &subCategoryID, &subCategoryName, &name, &shortDescription, &description, &mainPurpose, &features, &formulaRef, &erpID)
		category := NewCategory(categoryID, cateogryName)
		product := NewProduct(id, category, subCategoryID, subCategoryName, name, shortDescription, description, mainPurpose, features, formulaRef, erpID)
		products = append(products, product)
	}
	return products, nil
}
