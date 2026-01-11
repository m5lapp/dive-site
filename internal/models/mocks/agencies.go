package mocks

import "github.com/m5lapp/divesite-monolith/internal/models"

var agencyBSAC = models.Agency{
	ID:         1,
	CommonName: "PADI",
	FullName:   "Professional Association of Diving Instructors",
	Acronym:    "PADI",
	URL:        "https://www.padi.com/",
}

type AgencyModel struct{}

func (m *AgencyModel) List() ([]models.Agency, error) {
	return []models.Agency{agencyBSAC}, nil
}

var agencyCourseBSACOceanDiver = models.AgencyCourse{
	ID:                1,
	Agency:            agencyBSAC,
	Name:              "Ocean Diver",
	URL:               "https://www.bsac.com/page.asp?section=3303&sectionTitle=Ocean+Diver",
	IsSpecialtyCourse: false,
	IsTechCourse:      false,
	IsProCourse:       false,
}

type AgencyCourseModel struct{}

func (m *AgencyCourseModel) List() ([]models.AgencyCourse, error) {
	return []models.AgencyCourse{agencyCourseBSACOceanDiver}, nil
}
