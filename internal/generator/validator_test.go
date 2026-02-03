package generator

import (
	"strings"
	"testing"
)

func TestNewValidator(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v, want nil", err)
	}
	if v == nil {
		t.Fatal("NewValidator() returned nil validator")
	}
	if v.schema == nil {
		t.Fatal("NewValidator() returned validator with nil schema")
	}
}

func TestValidator_Validate(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "empty object is valid",
			input:   `{}`,
			wantErr: false,
		},
		{
			name: "valid minimal JSON Resume - basics only",
			input: `{
				"basics": {
					"name": "John Doe",
					"email": "john@example.com"
				}
			}`,
			wantErr: false,
		},
		{
			name: "valid full JSON Resume - basics + work + education",
			input: `{
				"basics": {
					"name": "Jane Smith",
					"label": "Software Engineer",
					"email": "jane@example.com",
					"phone": "+1-555-1234",
					"summary": "Experienced software engineer with 10+ years of experience."
				},
				"work": [
					{
						"name": "Tech Company",
						"position": "Senior Developer",
						"startDate": "2020-01",
						"endDate": "2024-06",
						"summary": "Led development of key features",
						"highlights": ["Increased performance by 40%"]
					}
				],
				"education": [
					{
						"institution": "State University",
						"area": "Computer Science",
						"studyType": "Bachelor",
						"startDate": "2010-09",
						"endDate": "2014-05"
					}
				]
			}`,
			wantErr: false,
		},
		{
			name: "valid with all sections",
			input: `{
				"basics": {"name": "Test User"},
				"work": [],
				"education": [],
				"awards": [],
				"certificates": [],
				"publications": [],
				"skills": [],
				"languages": [],
				"interests": [],
				"references": [],
				"projects": []
			}`,
			wantErr: false,
		},
		{
			name: "valid with location",
			input: `{
				"basics": {
					"name": "John Doe",
					"location": {
						"city": "San Francisco",
						"region": "California",
						"countryCode": "US"
					}
				}
			}`,
			wantErr: false,
		},
		{
			name: "valid with profiles",
			input: `{
				"basics": {
					"name": "John Doe",
					"profiles": [
						{
							"network": "LinkedIn",
							"username": "johndoe"
						},
						{
							"network": "GitHub",
							"username": "johndoe"
						}
					]
				}
			}`,
			wantErr: false,
		},
		{
			name: "valid with skills and keywords",
			input: `{
				"skills": [
					{
						"name": "Web Development",
						"level": "Expert",
						"keywords": ["HTML", "CSS", "JavaScript"]
					}
				]
			}`,
			wantErr: false,
		},
		{
			name: "invalid - basics.email wrong type (number instead of string)",
			input: `{
				"basics": {
					"name": "John Doe",
					"email": 12345
				}
			}`,
			wantErr:     true,
			errContains: "email",
		},
		{
			name: "invalid - work should be array not object",
			input: `{
				"work": {
					"name": "Company"
				}
			}`,
			wantErr:     true,
			errContains: "work",
		},
		{
			name: "invalid - education should be array not string",
			input: `{
				"education": "MIT"
			}`,
			wantErr:     true,
			errContains: "education",
		},
		{
			name: "invalid - skills items should be objects",
			input: `{
				"skills": ["JavaScript", "Python"]
			}`,
			wantErr:     true,
			errContains: "skills",
		},
		{
			name:        "invalid JSON - completely broken",
			input:       `{not json at all`,
			wantErr:     true,
			errContains: "invalid JSON",
		},
		{
			name:        "invalid JSON - missing closing brace",
			input:       `{"name": "test"`,
			wantErr:     true,
			errContains: "invalid JSON",
		},
		{
			name: "valid - date formats accepted",
			input: `{
				"work": [
					{
						"name": "Company",
						"startDate": "2020-01-15",
						"endDate": "2024-06"
					}
				],
				"education": [
					{
						"institution": "University",
						"startDate": "2010"
					}
				]
			}`,
			wantErr: false,
		},
		{
			name: "invalid - date format wrong",
			input: `{
				"work": [
					{
						"name": "Company",
						"startDate": "January 2020"
					}
				]
			}`,
			wantErr:     true,
			errContains: "startDate",
		},
		{
			name: "valid - additional properties allowed",
			input: `{
				"basics": {
					"name": "John Doe",
					"customField": "custom value"
				},
				"customSection": {"data": "allowed"}
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.Validate([]byte(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() error = nil, wantErr = true")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Validate() error = %q, want error containing %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("Validate() error = %v, wantErr = false", err)
			}
		})
	}
}

func TestValidator_ValidateRealisticResume(t *testing.T) {
	v, err := NewValidator()
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}

	// Test with a realistic full resume
	realisticResume := `{
		"basics": {
			"name": "Richard Roe",
			"label": "Senior Software Engineer",
			"email": "richard.roe@example.com",
			"phone": "+1-555-123-4567",
			"url": "https://richardroe.dev",
			"summary": "Senior software engineer with 8+ years of experience in full-stack development. Specialized in Go, TypeScript, and cloud infrastructure. Passionate about building scalable systems and mentoring junior developers.",
			"location": {
				"city": "San Francisco",
				"region": "California",
				"countryCode": "US"
			},
			"profiles": [
				{
					"network": "GitHub",
					"username": "richardroe",
					"url": "https://github.com/richardroe"
				},
				{
					"network": "LinkedIn",
					"username": "richardroe",
					"url": "https://linkedin.com/in/richardroe"
				}
			]
		},
		"work": [
			{
				"name": "TechCorp Inc.",
				"position": "Senior Software Engineer",
				"startDate": "2020-03",
				"summary": "Leading backend development for core platform services.",
				"highlights": [
					"Designed and implemented microservices architecture serving 10M+ daily requests",
					"Reduced API latency by 60% through query optimization",
					"Mentored team of 4 junior engineers"
				]
			},
			{
				"name": "StartupXYZ",
				"position": "Software Engineer",
				"startDate": "2017-06",
				"endDate": "2020-02",
				"summary": "Full-stack development for B2B SaaS platform.",
				"highlights": [
					"Built real-time collaboration features using WebSockets",
					"Implemented CI/CD pipeline reducing deployment time by 80%"
				]
			}
		],
		"education": [
			{
				"institution": "University of California, Berkeley",
				"area": "Computer Science",
				"studyType": "Bachelor of Science",
				"startDate": "2013-08",
				"endDate": "2017-05",
				"score": "3.8/4.0"
			}
		],
		"skills": [
			{
				"name": "Backend Development",
				"level": "Expert",
				"keywords": ["Go", "Python", "Node.js", "PostgreSQL", "Redis"]
			},
			{
				"name": "Frontend Development",
				"level": "Advanced",
				"keywords": ["TypeScript", "React", "Vue.js", "HTML/CSS"]
			},
			{
				"name": "Cloud & DevOps",
				"level": "Advanced",
				"keywords": ["AWS", "Kubernetes", "Docker", "Terraform"]
			}
		],
		"languages": [
			{
				"language": "English",
				"fluency": "Native"
			},
			{
				"language": "Spanish",
				"fluency": "Intermediate"
			}
		],
		"projects": [
			{
				"name": "Open Source CLI Tool",
				"description": "A command-line tool for automating development workflows",
				"highlights": ["1000+ GitHub stars", "Used by 50+ companies"],
				"keywords": ["Go", "CLI", "Open Source"],
				"url": "https://github.com/richardroe/cli-tool"
			}
		]
	}`

	if err := v.Validate([]byte(realisticResume)); err != nil {
		t.Errorf("Validate() realistic resume error = %v", err)
	}
}
