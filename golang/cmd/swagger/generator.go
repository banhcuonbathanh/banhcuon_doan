// cmd/swagger/generator.go
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type SwaggerSpec struct {
	Swagger      string                 `yaml:"swagger"`
	Info         map[string]interface{} `yaml:"info"`
	Host         string                 `yaml:"host"`
	BasePath     string                 `yaml:"basePath"`
	Schemes      []string               `yaml:"schemes"`
	Consumes     []string               `yaml:"consumes"`
	Produces     []string               `yaml:"produces"`
	Tags         []map[string]interface{} `yaml:"tags"`
	SecurityDefinitions map[string]interface{} `yaml:"securityDefinitions"`
	Paths        map[string]interface{} `yaml:"paths"`
	Definitions  map[string]interface{} `yaml:"definitions"`
	Responses    map[string]interface{} `yaml:"responses"`
	ExternalDocs map[string]interface{} `yaml:"externalDocs"`
}

type PathSpec struct {
	Paths map[string]interface{} `yaml:",inline"`
}

type DefinitionSpec struct {
	Definitions map[string]interface{} `yaml:"definitions"`
}

func main() {
	fmt.Println("🚀 Generating comprehensive Swagger documentation...")

	// Define paths
	docsPath := "docs/swagger"
	outputPath := "docs/generated"
	outputFile := filepath.Join(outputPath, "swagger.yaml")

	// Create output directory
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Load main swagger configuration
	mainSpec, err := loadMainSpec(filepath.Join(docsPath, "main.yaml"))
	if err != nil {
		log.Fatalf("Failed to load main spec: %v", err)
	}

	// Load definitions
	definitions, err := loadDefinitions(filepath.Join(docsPath, "definitions/common.yaml"))
	if err != nil {
		log.Fatalf("Failed to load definitions: %v", err)
	}
	mainSpec.Definitions = definitions

	// Load account auth endpoints
	authPaths, err := loadAccountAuthPaths(filepath.Join(docsPath, "account/auth"))
	if err != nil {
		log.Fatalf("Failed to load auth paths: %v", err)
	}
	mainSpec.Paths = authPaths

	// Generate final swagger file
	if err := generateSwaggerFile(mainSpec, outputFile); err != nil {
		log.Fatalf("Failed to generate swagger file: %v", err)
	}

	fmt.Printf("✅ Swagger documentation generated successfully!\n")
	fmt.Printf("📄 Output file: %s\n", outputFile)
	fmt.Printf("🌐 You can now use this file with Swagger UI or other API documentation tools\n")
	
	// Generate additional formats
	if err := generateAdditionalFormats(mainSpec, outputPath); err != nil {
		log.Printf("⚠️  Warning: Failed to generate additional formats: %v", err)
	}
}

func loadMainSpec(filePath string) (*SwaggerSpec, error) {
	fmt.Printf("📖 Loading main specification from %s\n", filePath)
	
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read main spec file: %w", err)
	}

	var spec SwaggerSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse main spec YAML: %w", err)
	}

	return &spec, nil
}

func loadDefinitions(filePath string) (map[string]interface{}, error) {
	fmt.Printf("📖 Loading definitions from %s\n", filePath)
	
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read definitions file: %w", err)
	}

	var defSpec DefinitionSpec
	if err := yaml.Unmarshal(data, &defSpec); err != nil {
		return nil, fmt.Errorf("failed to parse definitions YAML: %w", err)
	}

	return defSpec.Definitions, nil
}

func loadAccountAuthPaths(authDir string) (map[string]interface{}, error) {
	fmt.Printf("📖 Loading account auth paths from %s\n", authDir)
	
	paths := make(map[string]interface{})
	
	// List of auth endpoint files
	authFiles := []string{"login.yaml", "register.yaml", "logout.yaml"}
	
	for _, filename := range authFiles {
		filePath := filepath.Join(authDir, filename)
		
		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Printf("⚠️  Warning: Auth file not found: %s\n", filePath)
			continue
		}
		
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read auth file %s: %w", filename, err)
		}

		var pathSpec PathSpec
		if err := yaml.Unmarshal(data, &pathSpec); err != nil {
			return nil, fmt.Errorf("failed to parse auth YAML %s: %w", filename, err)
		}

		// Merge paths
		for path, spec := range pathSpec.Paths {
			paths[path] = spec
			fmt.Printf("  ✅ Loaded endpoint: %s\n", path)
		}
	}

	return paths, nil
}

func generateSwaggerFile(spec *SwaggerSpec, outputFile string) error {
	fmt.Printf("🔧 Generating final Swagger file: %s\n", outputFile)
	
	data, err := yaml.Marshal(spec)
	if err != nil {
		return fmt.Errorf("failed to marshal swagger spec: %w", err)
	}

	if err := ioutil.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write swagger file: %w", err)
	}

	return nil
}

func generateAdditionalFormats(spec *SwaggerSpec, outputPath string) error {
	// Generate JSON format
	fmt.Printf("🔧 Generating JSON format...\n")
	
	// Convert YAML to JSON (simplified - you might want to use a proper converter)
	jsonFile := filepath.Join(outputPath, "swagger.json")
	
	// Create a simple JSON version indicator file
	jsonContent := `{
  "info": "This is a placeholder. Use yaml-to-json converter to generate swagger.json from swagger.yaml",
  "source": "swagger.yaml"
}`
	
	if err := ioutil.WriteFile(jsonFile, []byte(jsonContent), 0644); err != nil {
		return fmt.Errorf("failed to write JSON placeholder: %w", err)
	}

	// Generate documentation summary
	fmt.Printf("🔧 Generating documentation summary...\n")
	
	summaryFile := filepath.Join(outputPath, "README.md")
	summary := generateDocumentationSummary(spec)
	
	if err := ioutil.WriteFile(summaryFile, []byte(summary), 0644); err != nil {
		return fmt.Errorf("failed to write documentation summary: %w", err)
	}

	return nil
}

func generateDocumentationSummary(spec *SwaggerSpec) string {
	var sb strings.Builder
	
	sb.WriteString("# API Documentation Summary\n\n")
	sb.WriteString("This document was automatically generated from the Swagger specification.\n\n")
	
	if info, ok := spec.Info["title"].(string); ok {
		sb.WriteString(fmt.Sprintf("## %s\n\n", info))
	}
	
	if desc, ok := spec.Info["description"].(string); ok {
		sb.WriteString(fmt.Sprintf("%s\n\n", desc))
	}
	
	sb.WriteString("## Available Endpoints\n\n")
	
	for path := range spec.Paths {
		sb.WriteString(fmt.Sprintf("- `%s`\n", path))
	}
	
	sb.WriteString("\n## Files Generated\n\n")
	sb.WriteString("- `swagger.yaml` - Complete OpenAPI/Swagger specification\n")
	sb.WriteString("- `swagger.json` - JSON format (use converter)\n")
	sb.WriteString("- `README.md` - This summary file\n\n")
	
	sb.WriteString("## Usage\n\n")
	sb.WriteString("1. **Swagger UI**: Upload `swagger.yaml` to https://editor.swagger.io/\n")
	sb.WriteString("2. **Local Server**: Use swagger-ui-dist with the generated files\n")
	sb.WriteString("3. **API Testing**: Import into Postman, Insomnia, or similar tools\n\n")
	
	sb.WriteString("## Test Scenarios Included\n\n")
	sb.WriteString("Each endpoint includes comprehensive test scenarios:\n")
	sb.WriteString("- ✅ Success cases with example responses\n")
	sb.WriteString("- ❌ Error cases with specific error codes\n")
	sb.WriteString("- 📝 Request/response examples\n")
	sb.WriteString("- 🔧 Code samples in multiple languages\n\n")
	
	return sb.String()
}