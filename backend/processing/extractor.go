// backend/processing/extractor.go
package processing

import (
	"bytes"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	/*"github.com/malharg/strategic-insight-analyst/backend/config"
	"github.com/unidoc/unipdf/v3/common/license"*/
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

/*func init() {
	// This init function runs once when the package is first used.
	// It sets the UniDoc license key for the entire application.
	err := license.SetMeteredKey(config.AppConfig.UnidocLicenseKey)
	if err != nil {
		log.Fatalf("FATAL: Failed to set UniDoc license key: %v", err)
	}
	log.Println("UniDoc license key set successfully.")
}*/

// ExtractTextFromFile uses UniDoc for PDFs.
func ExtractTextFromFile(fileBytes []byte, fileName string) (string, error) {
	extension := strings.ToLower(filepath.Ext(fileName))

	switch extension {
	case ".txt":
		return string(fileBytes), nil
	case ".pdf":
		return extractTextFromPDF(fileBytes)
	default:
		return "", fmt.Errorf("unsupported file type: %s", extension)
	}
}

// extractTextFromPDF uses the UniDoc library.
func extractTextFromPDF(fileBytes []byte) (string, error) {
	// Create a new PDF reader from the file bytes.
	pdfReader, err := model.NewPdfReader(bytes.NewReader(fileBytes))
	if err != nil {
		log.Printf("ERROR: UniDoc failed to create PDF reader: %v", err)
		return "", err
	}

	// Get the total number of pages in the PDF.
	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		log.Printf("ERROR: UniDoc failed to get page count: %v", err)
		return "", err
	}

	// Extract text from all pages and concatenate.
	var extractedText strings.Builder
	for i := 1; i <= numPages; i++ {
		page, err := pdfReader.GetPage(i)
		if err != nil {
			log.Printf("ERROR: UniDoc failed to get page %d: %v", i, err)
			return "", err
		}

		ex, err := extractor.New(page)
		if err != nil {
			log.Printf("ERROR: UniDoc failed to create extractor for page %d: %v", i, err)
			return "", err
		}

		text, err := ex.ExtractText()
		if err != nil {
			log.Printf("ERROR: UniDoc failed to extract text from page %d: %v", i, err)
			// Continue to the next page even if one fails
			continue
		}

		extractedText.WriteString(text)
		extractedText.WriteString("\n\n") // Add a separator between pages
	}

	return extractedText.String(), nil
}
