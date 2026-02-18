package database

import (
	"go-auth-api/internal/models"
	"gorm.io/gorm"
)

func SeedRolesAndPermissions(db *gorm.DB) {
	// 1. Define Permissions
	perms := []models.Permission{
		{Name: "View Invoices", Slug: "FIN_VIEW"},
		{Name: "Manage Employees", Slug: "HR_MANAGE"},
		{Name: "System Config", Slug: "ICT_ADMIN"},
	}
	for _, p := range perms {
		db.FirstOrCreate(&p, models.Permission{Slug: p.Slug})
	}

	// 2. Define Roles and Map Permissions
	roles := []struct {
		Name  string
		Perms []string
	}{
		{"Finance", []string{"FIN_VIEW"}},
		{"HR", []string{"HR_MANAGE"}},
		{"ICT", []string{"ICT_ADMIN", "FIN_VIEW", "HR_MANAGE"}},
	}

	for _, r := range roles {
		var role models.Role
		db.FirstOrCreate(&role, models.Role{Name: r.Name})
		
		var pList []models.Permission
		db.Where("slug IN ?", r.Perms).Find(&pList)
		db.Model(&role).Association("Permissions").Replace(pList)
	}
}