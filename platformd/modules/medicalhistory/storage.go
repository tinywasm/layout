package medicalhistory

// Patient and Visit are LOCAL FAKE types for this demo only. A real
// deployment replaces todayAgenda/visits below with router.Caller calls
// into appointment_booking (today's reservations for a staff member) and
// clinical_encounter (ListVisitsByPatient) — this repo never imports either.
type Patient struct {
	ID   string
	Name string
	Time string
}

type Visit struct {
	Date      string
	Reason    string
	Diagnosis string
}

type patientVisit struct {
	PatientID string
	Visit     Visit
}

// todayAgenda is the schedule for today, with exactly 2 patients who reserved for the day according to the hour.
var todayAgenda = []Patient{
	{ID: "p1", Name: "Juan Pérez", Time: "09:00"},
	{ID: "p2", Name: "María Soto", Time: "10:15"},
}

// visits has exactly 1 medical history record for each of the 2 patients.
var visits = []patientVisit{
	{"p1", Visit{Date: "2026-07-20", Reason: "Control", Diagnosis: "Sin hallazgos"}},
	{"p2", Visit{Date: "2026-06-02", Reason: "Chequeo anual", Diagnosis: "Saludable"}},
}

func historyFor(patientID string) []Visit {
	var out []Visit
	for _, pv := range visits {
		if pv.PatientID == patientID {
			out = append(out, pv.Visit)
		}
	}
	return out
}
