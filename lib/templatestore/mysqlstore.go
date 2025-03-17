package templatestore

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

// MySQLTemplateStorage implements the StorageAdapter interface using MySQL as backend.
type MySQLTemplateStorage struct {
	DB *sql.DB
}

// NewMySQLStorageAdapter creates and initializes a new MySQL storage adapter.
func NewMySQLStorageAdapter(dsn string) (*MySQLTemplateStorage, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL: %v", err)
	}

	storage := &MySQLTemplateStorage{
		DB: db,
	}

	// Initialize the database
	if err := storage.initDatabase(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	return storage, nil
}

// initDatabase creates the necessary tables if they don't exist.
func (m *MySQLTemplateStorage) initDatabase() error {
	// Check if templates table exists using information_schema
	var count int
	err := m.DB.QueryRow(`
		 SELECT COUNT(*) 
		 FROM information_schema.tables 
		 WHERE table_schema = DATABASE() 
		 AND table_name = 'templates'`).Scan(&count)

	if err != nil {
		return fmt.Errorf("failed to check for templates table: %v", err)
	}

	if count == 0 {
		// Table doesn't exist
		return fmt.Errorf("templates table doesn't exist in the database - please run the initialization script first")
	}
	// Validate table structure
	var templateIdCol, templateContentCol, templateNameCol int
	err = m.DB.QueryRow(`
			SELECT 
				COUNT(*) FROM information_schema.columns 
				WHERE table_schema = DATABASE() 
				AND table_name = 'templates' 
				AND column_name = 'template_id'`).Scan(&templateIdCol)
	if err != nil || templateIdCol == 0 {
		return fmt.Errorf("templates table is missing template_id column")
	}

	err = m.DB.QueryRow(`
			SELECT 
				COUNT(*) FROM information_schema.columns 
				WHERE table_schema = DATABASE() 
				AND table_name = 'templates' 
				AND column_name = 'template_content'`).Scan(&templateContentCol)
	if err != nil || templateContentCol == 0 {
		return fmt.Errorf("templates table is missing template_content column")
	}

	err = m.DB.QueryRow(`
			SELECT 
				COUNT(*) FROM information_schema.columns 
				WHERE table_schema = DATABASE() 
				AND table_name = 'templates' 
				AND column_name = 'template_name'`).Scan(&templateNameCol)
	if err != nil || templateNameCol == 0 {
		return fmt.Errorf("templates table is missing template_name column")
	}

	err = m.DB.QueryRow(`
	SELECT 
		COUNT(*) FROM information_schema.columns 
		WHERE table_schema = DATABASE() 
		AND table_name = 'templates' 
		AND column_name = 'template_name'`).Scan(&templateNameCol)
	if err != nil || templateNameCol == 0 {
		return fmt.Errorf("templates table is missing template_name column")
	}

	return nil
}

// GetTemplate retrieves a template from MySQL.
func (m *MySQLTemplateStorage) GetTemplate(ctx context.Context, req *GetTemplateRequest) (*template.Template, error) {
	var templateContent string
	var templateID string

	if req.TemplateUUID == "" {
		return nil, fmt.Errorf("template UUID is required for MySQL storage")
	}

	err := m.DB.QueryRowContext(ctx, "SELECT template_content FROM templates WHERE template_id = ?", req.TemplateUUID).Scan(&templateContent)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template not found: %s", req.TemplateUUID)
		}
		return nil, fmt.Errorf("error retrieving template: %v", err)
	}

	return template.New("template" + templateID).Parse(templateContent)
}

// PutDocument stores a document in MySQL.
func (m *MySQLTemplateStorage) PutDocument(ctx context.Context, req *PostDocumentRequest, reader *io.Reader) (string, error) {
	return "", fmt.Errorf("put document not implemented for mysql, use other adapters for filestorage")
}

// GetDocument retrieves a document from MySQL.
func (m *MySQLTemplateStorage) GetDocument(ctx context.Context, req *GetDocumentRequest) (io.Reader, error) {
	return nil, fmt.Errorf("get document not implemented for mysql, use other adapters for filestorage")
}

// ListTemplates retrieves all templates from MySQL storage.
func (m *MySQLTemplateStorage) ListTemplates(ctx context.Context) ([]*TemplateInfo, error) {
	// Query all templates
	rows, err := m.DB.QueryContext(ctx, "SELECT template_id, template_name, created_at, updated_at FROM templates")
	if err != nil {
		return nil, fmt.Errorf("error querying templates: %v", err)
	}
	defer rows.Close()

	var templates []*TemplateInfo
	for rows.Next() {
		var template TemplateInfo
		var createdAt, updatedAt sql.NullTime

		if err := rows.Scan(&template.TemplateID, &template.TemplateName, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("error scanning template row: %v", err)
		}

		if createdAt.Valid {
			template.CreatedAt = createdAt.Time
		}

		if updatedAt.Valid {
			template.UpdatedAt = updatedAt.Time
		}

		templates = append(templates, &template)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating template rows: %v", err)
	}

	return templates, nil
}

// Helper function to get template from MySQL
func (m *MySQLTemplateStorage) GetTemplateContent(ctx context.Context, req *GetTemplateContentRequest) (*GetTemplateContentResponse, error) {
	// Query template info from database
	var templateContent, templateName, jsonSchema string
	templateId := req.TemplateUUID
	// First get the template content
	err := m.DB.QueryRowContext(ctx,
		"SELECT template_content, template_name,json_schema FROM templates WHERE template_id = ?",
		templateId).Scan(&templateContent, &templateName, &jsonSchema)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template not found: %s", templateId)
		}
		return nil, fmt.Errorf("error retrieving template: %v", err)
	}

	resp := &GetTemplateContentResponse{
		TemplateContent:    templateContent,
		TemplateName:       templateName,
		TemplateJsonSchema: jsonSchema,
	}

	return resp, nil
}

// CreateTemplate stores a template in the MySQL database
func (m *MySQLTemplateStorage) CreateTemplate(ctx context.Context, req *CreateTemplateRequest) (string, error) {
	// create a new UUID for the template ID
	templateID := uuid.New().String()
	// Check if template ID already exists
	var count int
	err := m.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM templates WHERE template_id = ?", templateID).Scan(&count)
	if err != nil {
		return "", fmt.Errorf("error checking for duplicate template: %v", err)
	}

	if count > 0 {
		fmt.Println("template ID already exists: ", templateID)
		return "", fmt.Errorf("server error, failed at id generation")
	}

	// Insert the template
	_, err = m.DB.ExecContext(ctx,
		"INSERT INTO templates (template_id, template_name, template_content, json_schema) VALUES (?, ?, ?, ?)",
		templateID, req.TemplateName, req.TemplateHTML, req.TemplateJSON)

	if err != nil {
		return "", fmt.Errorf("error inserting template into database: %v", err)
	}

	return templateID, nil
}

// Close closes the database connection.
func (m *MySQLTemplateStorage) Close() error {
	if m.DB != nil {
		return m.DB.Close()
	}
	return nil
}
