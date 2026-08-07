package handler

import (
	"net/http"

	"eregen.dev/api-server/internal/store"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MedicalHandler struct {
	pg *store.Postgres
	log *zap.Logger
}

func NewMedicalHandler(pg *store.Postgres, log *zap.Logger) *MedicalHandler {
	return &MedicalHandler{pg: pg, log: log}
}

// GET /api/v1/medical/patients/:patient_id/history
func (h *MedicalHandler) GetPatientHistory(c *gin.Context) {
	patientID := c.Param("patient_id")
	if patientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_REQUEST", "message": "patient_id is required"})
		return
	}

	// Query directly via DB pool since medical tables are SQLite-specific
	query := `
		SELECT p.id, p.hospital_id, p.name, p.gender, p.age, p.admission_no,
		       p.department, p.bed_no, p.blood_type, p.allergy_history,
		       p.special_disease, p.alert_tags, p.medical_summary, p.status,
		       p.created_at, p.updated_at
		FROM medical_wb_patients p WHERE p.id = ?`
	var id, hospitalID, name, admissionNo, department, bedNo, bloodType, allergyHist, specialDisease, alertTags, medSummary, status string
	var gender, age int
	var createdAt, updatedAt string
	err := h.pg.Pool().QueryRow(c.Request.Context(), query, patientID).Scan(
		&id, &hospitalID, &name, &gender, &age, &admissionNo,
		&department, &bedNo, &bloodType, &allergyHist,
		&specialDisease, &alertTags, &medSummary, &status,
		&createdAt, &updatedAt,
	)
	if err != nil {
		h.log.Error("get patient", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Patient not found"})
		return
	}

	patient := gin.H{
		"id": id, "hospital_id": hospitalID, "name": name, "gender": gender,
		"age": age, "admission_no": admissionNo, "department": department,
		"bed_no": bedNo, "blood_type": bloodType, "allergy_history": allergyHist,
		"special_disease": specialDisease, "alert_tags": alertTags,
		"medical_summary": medSummary, "status": status,
		"created_at": createdAt, "updated_at": updatedAt,
	}

	// Expenses
	expenses, _ := h.QueryExpenses(c, patientID)
	// Medications
	meds, _ := h.QueryMedications(c, patientID)
	// Test results
	tests, _ := h.QueryTestResults(c, patientID)
	// Daily entries
	entries, _ := h.QueryDailyEntries(c, patientID, "")
	// Verifications
	verifs, _ := h.QueryVerifications(c, patientID)

	c.JSON(http.StatusOK, gin.H{
		"code": "OK", "data": gin.H{
			"patient": patient,
			"expenses": expenses,
			"medications": meds,
			"test_results": tests,
			"daily_entries": entries,
			"verifications": verifs,
		},
	})
}

func (h *MedicalHandler) QueryExpenses(c *gin.Context, patientID string) ([]gin.H, error) {
	rows, err := h.pg.Pool().Query(c.Request.Context(),
		`SELECT id, item_name, item_type, amount, quantity, unit_price,
		        COALESCE(recorded_date,''), COALESCE(recorded_by,''), created_at
		 FROM medical_wb_expenses WHERE patient_id = ? ORDER BY recorded_date DESC`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []gin.H
	for rows.Next() {
		var item gin.H
		var id, itemName, itemType, recordedDate, recordedBy, createdAt string
		var amount, quantity, unitPrice float64
		if err := rows.Scan(&id, &itemName, &itemType, &amount, &quantity, &unitPrice, &recordedDate, &recordedBy, &createdAt); err != nil {
			return nil, err
		}
		item = gin.H{"id": id, "item_name": itemName, "item_type": itemType, "amount": amount,
			"quantity": quantity, "unit_price": unitPrice, "recorded_date": recordedDate,
			"recorded_by": recordedBy, "created_at": createdAt}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (h *MedicalHandler) QueryMedications(c *gin.Context, patientID string) ([]gin.H, error) {
	rows, err := h.pg.Pool().Query(c.Request.Context(),
		`SELECT id, drug_name, COALESCE(dosage,''), COALESCE(frequency,''),
		        COALESCE(route,''), COALESCE(start_date,''), COALESCE(end_date,''),
		        COALESCE(status,'active'), COALESCE(recorded_by,''), created_at
		 FROM medical_wb_medications WHERE patient_id = ? ORDER BY created_at DESC`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []gin.H
	for rows.Next() {
		var item gin.H
		var id, drugName, dosage, freq, route, startDate, endDate, status, recordedBy, createdAt string
		if err := rows.Scan(&id, &drugName, &dosage, &freq, &route, &startDate, &endDate, &status, &recordedBy, &createdAt); err != nil {
			return nil, err
		}
		item = gin.H{"id": id, "drug_name": drugName, "dosage": dosage, "frequency": freq,
			"route": route, "start_date": startDate, "end_date": endDate,
			"status": status, "recorded_by": recordedBy, "created_at": createdAt}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (h *MedicalHandler) QueryTestResults(c *gin.Context, patientID string) ([]gin.H, error) {
	rows, err := h.pg.Pool().Query(c.Request.Context(),
		`SELECT id, test_name, COALESCE(test_date,''), COALESCE(result_text,''),
		        COALESCE(result_value,''), COALESCE(reference_range,''), abnormal_flag,
		        COALESCE(recorded_by,''), created_at
		 FROM medical_wb_test_results WHERE patient_id = ? ORDER BY test_date DESC`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []gin.H
	for rows.Next() {
		var item gin.H
		var id, testName, testDate, resultText, resultValue, refRange, recordedBy, createdAt string
		var abnormalFlag int
		if err := rows.Scan(&id, &testName, &testDate, &resultText, &resultValue, &refRange, &abnormalFlag, &recordedBy, &createdAt); err != nil {
			return nil, err
		}
		item = gin.H{"id": id, "test_name": testName, "test_date": testDate,
			"result_text": resultText, "result_value": resultValue, "reference_range": refRange,
			"abnormal_flag": abnormalFlag, "recorded_by": recordedBy, "created_at": createdAt}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (h *MedicalHandler) QueryDailyEntries(c *gin.Context, patientID, date string) ([]gin.H, error) {
	query := `SELECT id, entry_date, entry_type, content, COALESCE(recorded_by,''), created_at
		FROM medical_wb_daily_entries WHERE patient_id = ?`
	args := []interface{}{patientID}
	if date != "" {
		query += " AND entry_date = ?"
		args = append(args, date)
	}
	query += " ORDER BY entry_date DESC"
	rows, err := h.pg.Pool().Query(c.Request.Context(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []gin.H
	for rows.Next() {
		var item gin.H
		var id, entryDate, entryType, content, recordedBy, createdAt string
		if err := rows.Scan(&id, &entryDate, &entryType, &content, &recordedBy, &createdAt); err != nil {
			return nil, err
		}
		item = gin.H{"id": id, "entry_date": entryDate, "entry_type": entryType,
			"content": content, "recorded_by": recordedBy, "created_at": createdAt}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (h *MedicalHandler) QueryVerifications(c *gin.Context, patientID string) ([]gin.H, error) {
	rows, err := h.pg.Pool().Query(c.Request.Context(),
		`SELECT id, device_id, COALESCE(nurse_id,''), action, verified_at,
		        COALESCE(read_data,''), COALESCE(result,'success'), COALESCE(notes,'')
		 FROM medical_wb_verifications WHERE patient_id = ? ORDER BY verified_at DESC`, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []gin.H
	for rows.Next() {
		var item gin.H
		var id, deviceID, nurseID, action, verifiedAt, readData, result, notes string
		if err := rows.Scan(&id, &deviceID, &nurseID, &action, &verifiedAt, &readData, &result, &notes); err != nil {
			return nil, err
		}
		item = gin.H{"id": id, "device_id": deviceID, "nurse_id": nurseID, "action": action,
			"verified_at": verifiedAt, "read_data": readData, "result": result, "notes": notes}
		items = append(items, item)
	}
	return items, rows.Err()
}
