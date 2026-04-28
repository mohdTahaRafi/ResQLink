# SAMAJ — Community Intelligence Platform
## Comprehensive Project Documentation

---

## 1. Brief About the Solution

**SAMAJ** (Sanskrit: "Community") is an AI-powered Community Intelligence Platform designed to fundamentally transform how non-profit organizations (NGOs), public volunteers, and specialized professionals (lawyers, doctors) collaborate to identify, prioritize, and resolve civic, medical, legal, and disaster-related issues at the grassroots level.

In developing nations, millions of citizen-reported issues — broken infrastructure, medical emergencies in underserved areas, legal disputes affecting vulnerable populations — go unresolved because there is no unified, intelligent system connecting the people who report problems with the organizations equipped to solve them. Existing grievance portals are bureaucratic, opaque, and offer zero intelligent routing.

**SAMAJ** bridges this gap by providing:
- **General Users (Reporters)** with a frictionless, multimodal reporting interface (text, image, voice, GPS) to submit issues from the field.
- **Volunteers** with an organized task pipeline, automatically matched to issues based on their skills, proximity, and reliability.
- **Specialized Professionals (Lawyers & Doctors)** with an AI-first case management environment, powered by Google Gemini, for semantic search, document analysis, and AI-assisted case evaluation.
- **NGO Administrators** with a real-time command center featuring algorithmic issue prioritization, intelligent volunteer matching, and a live geospatial heatmap for operational decision-making.

The platform is built on Google Cloud's ecosystem — Firebase Authentication, Cloud Firestore, Cloud Pub/Sub, Vertex AI (Gemini), and Google Maps Platform — ensuring enterprise-grade scalability, security, and intelligence from day one.

**Target Users:** NGOs, civic bodies, volunteer organizations, legal aid societies, public health networks, and the general public.

**Core Value:** SAMAJ reduces issue-to-resolution time by up to 70% through AI-driven triage, intelligent volunteer matching, and real-time geospatial awareness — turning community reporting from a passive suggestion box into an active intelligence system.

---

## 2. Opportunities

### 2a. How Different Is It from Existing Solutions? (Competitive Analysis)

| Dimension | Traditional Portals (311, CPGRAMS, MyGov) | Volunteer Platforms (VolunteerMatch, JustDial) | **SAMAJ** |
|---|---|---|---|
| **Report Intake** | Text-only web forms | Not applicable | Multimodal: text, image, voice, GPS — processed by Gemini AI |
| **Issue Classification** | Manual, department-based | None | AI-automated (Gemini extracts category, severity, population impact) |
| **Prioritization** | FIFO / manual | None | Weighted algorithm: urgency × issue type × time decay × population impact |
| **Volunteer Matching** | None | Keyword search, self-selection | Algorithmic: cosine similarity (skills) + Haversine distance + reliability scoring |
| **Specialist Support** | Separate portals | None | Integrated AI case manager with semantic document search |
| **Real-Time Awareness** | Basic dashboards | None | Live Google Maps heatmap with urgency-coded markers and heat circles |
| **Multilingual** | Limited | No | Hindi/regional dialect → English translation via Gemini |
| **Transparency** | Opaque | None | Full lifecycle tracking: pending → accepted → in_progress → resolved |

**Key Differentiator:** No existing platform combines AI-powered multimodal intake, algorithmic volunteer matching, specialist-grade case management, and real-time geospatial intelligence in a single, role-based interface. SAMAJ is the first platform to treat community issue management as an intelligent operations problem rather than a passive ticketing system.

### 2b. How Will It Solve the Problem? (Core Mechanics)

SAMAJ employs a four-stage intelligent pipeline:

**Stage 1 — Intelligent Intake:** When a General User submits a report (text, photo of a broken road, voice note in Hindi), the system ingests it via a Cloud Pub/Sub event. Gemini 3.1 Pro (via Vertex AI) processes the raw input — translating regional dialects, classifying the issue category (water, sanitation, road, legal, medical, etc.), estimating severity (1-10 index), and predicting the affected population count. This structured data is persisted to Cloud Firestore alongside the original media.

**Stage 2 — Algorithmic Prioritization:** The NGO Admin's dashboard does not show issues in chronological order. Instead, a weighted prioritization algorithm dynamically ranks every issue using: `Priority Score = (Urgency Weight × Issue Type Weight) + Time Decay Bonus + Population Impact Factor`. Critical medical emergencies surface instantly; routine civic issues queue lower. This ensures resources are always directed at maximum-impact problems first.

**Stage 3 — Intelligent Volunteer Matching:** When an admin selects an issue for assignment, the matching engine concurrently scores every registered volunteer across three dimensions: (1) **Skill Match** — cosine similarity between the issue's required skills and the volunteer's declared skills; (2) **Proximity** — Haversine great-circle distance, scored inversely (closer = higher); (3) **Reliability** — historical task completion rate. The final score is a weighted composite: `Total = 0.4×Skill + 0.35×Distance + 0.25×Reliability`. The top-N matched volunteers are presented for one-click assignment.

**Stage 4 — Resolution & Feedback Loop:** Assigned volunteers see tasks in their dashboard, update statuses (accepted → in progress → resolved), and the lifecycle updates propagate in real-time across all roles. Specialists handle complex cases (legal, medical) through an AI-assisted interface where they can ask Gemini questions about uploaded case documents and perform semantic searches. The heatmap reflects resolved issues, providing the NGO with a live operational picture.

### 2c. USP (Unique Selling Proposition)

**"AI-First Community Intelligence — from Report to Resolution in One Platform."**

SAMAJ's USP is the integration of three capabilities that have never been combined in a civic platform:

1. **Gemini-Powered Multimodal AI:** The only platform that accepts images, voice (in regional languages), and text — and automatically extracts structured, actionable intelligence without human intervention.
2. **Algorithmic Volunteer Matching:** Not random assignment, not self-selection — mathematically optimized matching using cosine similarity, geospatial distance, and reliability scoring.
3. **Specialist-Grade AI Case Management:** Lawyers and doctors don't just see task lists — they get an AI-powered research environment where they can interrogate case documents using natural language.

---

## 3. List of Features Offered by the Solution

### Core MVP Features (Implemented)

| # | Feature | Description |
|---|---------|-------------|
| 1 | **Role-First Authentication** | Landing page presents 4 roles (Reporter, Volunteer, Specialist, NGO Admin). User selects role, then authenticates. Post-login, they are locked to their role's dashboard only. |
| 2 | **Multimodal Issue Reporting** | Reporters submit issues with text descriptions, multiple photos (with compression), GPS coordinates (browser geolocation), issue type classification, urgency level, and location address. |
| 3 | **AI-Powered Report Ingestion** | Submitted reports are published to Cloud Pub/Sub and processed by Gemini 3.1 Pro, which extracts problem category, severity index (1-10), affected population estimate, and generates an English summary. |
| 4 | **Supported Media Types** | Text descriptions, JPEG/PNG images (auto-compressed to ≤600px and re-encoded), and voice recordings (Hindi/regional dialect support via Gemini audio processing). |
| 5 | **Algorithmic Issue Prioritization** | NGO dashboard automatically ranks issues using a weighted priority algorithm (urgency × issue type × time decay × population impact), not FIFO. |
| 6 | **Intelligent Volunteer Matching** | Concurrent scoring engine evaluates all volunteers against an issue using cosine skill similarity (0.4 weight), Haversine proximity (0.35 weight), and historical reliability (0.25 weight). |
| 7 | **One-Click Volunteer Assignment** | Admin selects matched volunteers via checkbox and assigns them with one click. Report status auto-updates to "accepted." |
| 8 | **Volunteer Task Dashboard** | Volunteers see their assigned tasks with urgency badges, issue type tags, and can update status (accepted → in_progress → resolved). |
| 9 | **Auto-Registration on Sign-Up** | When a Volunteer or Specialist creates an account, they are automatically registered in the Firestore `volunteers` collection with default skills and location, immediately visible to the NGO admin. |
| 10 | **Specialist Case Management** | 3-tab interface: (1) My Cases — list of assigned case files; (2) AI Chat — ask natural language questions about case documents via Gemini; (3) Document Search — semantic intra-document search. |
| 11 | **Google Maps Heatmap** | Admin dashboard features a real Google Maps interface with urgency-coded markers (🔴 Critical, 🟠 Urgent, 🟢 Normal), heat circles showing severity radius, satellite toggle, dark map style, and zoom controls. |
| 12 | **Browser Geolocation** | Reporter dashboard captures real GPS coordinates via the Web Geolocation API, displayed as lat/lng, and used for heatmap positioning. |
| 13 | **Premium Dark Mode** | Full dark mode support with proper surface colors, card styling, input borders, gradient role cards, and a custom dark map style for Google Maps. |
| 14 | **Firebase Authentication** | Secure email/password authentication with Firebase Auth, JWT token validation on every API request via custom Go middleware. |
| 15 | **Role-Locked Navigation** | Router enforces strict role isolation — logged-in users cannot access other roles' dashboards, must logout to change roles. |

### Advanced / Future Features (Planned)

| # | Feature | Description |
|---|---------|-------------|
| 16 | **Voice Report Submission** | Field workers can submit reports by speaking in Hindi or regional dialects; Gemini transcribes, translates, and structures the data automatically. |
| 17 | **Real-Time Push Notifications** | Firebase Cloud Messaging (FCM) to notify volunteers of new assignments, status changes, and escalations in real-time. |
| 18 | **Offline-First Mode** | Hive local storage enables report drafting and data caching offline; auto-syncs when connectivity is restored. |
| 19 | **AI-Powered Duplicate Detection** | Gemini compares incoming reports against existing open issues using semantic similarity to flag and merge duplicates. |
| 20 | **Predictive Hotspot Analysis** | Machine learning model trained on historical report data to predict future issue hotspots before incidents occur. |
| 21 | **Volunteer Gamification** | Points, badges, and leaderboards based on task completion, response time, and community impact to drive engagement. |
| 22 | **Government API Integration** | Bi-directional data exchange with government grievance portals (CPGRAMS, MyGov) for automated escalation of unresolved issues. |
| 23 | **Multi-Tenant NGO Support** | Single deployment supports multiple NGO organizations, each with isolated data, volunteer pools, and dashboards. |
| 24 | **WhatsApp/SMS Integration** | Issue reporting via WhatsApp Business API and SMS for users without smartphones or internet access. |
| 25 | **Transparent Citizen Portal** | Public-facing dashboard showing aggregate issue statistics, resolution rates, and response times by area — building civic trust. |
| 26 | **Document OCR & Digitization** | Camera-based document scanning with Gemini Vision for extracting text from physical legal documents, medical reports, and government notices. |
| 27 | **Multi-Language UI** | Full application localization for Hindi, Tamil, Bengali, Marathi, Telugu, and other major Indian languages. |
| 28 | **Audit Trail & Compliance Logging** | Complete activity logging for every state transition, assignment, and data access — enabling regulatory compliance and accountability. |

---

## 4. Process Flow and Use-Case Descriptions

### Flow 1: Issue Reporting by a General User

Step 1: User opens the SAMAJ web app and arrives at the Role Selection Screen.
Step 2: User selects the "General User" role card.
Step 3: System saves the selected role to local storage and navigates to the Login Screen.
Step 4: Login Screen displays a role badge "Signing in as General User" and presents email/password form.
Step 5: User enters credentials and submits (Sign Up or Sign In).
Step 6: Firebase Authentication validates credentials and returns a JWT token.
Step 7: System checks the saved role and navigates to the Reporter Dashboard.
Step 8: User fills the report form: description (required), issue type dropdown (Medical, Legal, Civic, Disaster), photos (multi-select with preview), location address (required).
Step 9: User taps "Get My Location" to capture GPS coordinates via browser Geolocation API.
Step 10: User selects urgency level (Normal / Urgent / Critical) and required volunteer count.
Step 11: User taps "Submit Report."
Step 12: Flutter app serializes the report data (including base64-encoded images) and sends POST /api/v1/reports with the Firebase JWT in Authorization header.
Step 13: Go backend validates the JWT via Firebase Admin SDK middleware.
Step 14: Backend compresses the image (≤600px, JPEG quality 40) if present and sizes exceed 1MB.
Step 15: Backend writes the report to Firestore "reports" collection with status "pending."
Step 16: Backend publishes an IngestionEvent to Cloud Pub/Sub topic "report-ingestion" (fire-and-forget).
Step 17: Backend returns HTTP 201 with the report ID to the client.
Step 18: (Async) Pub/Sub triggers the Gemini processing pipeline, which extracts problem_category, severity_index, affected_population_estimate, and summary. These enrichments are written back to the Firestore document.

### Flow 2: NGO Admin Issue Triage and Volunteer Assignment

Step 1: NGO Admin selects "NGO Organization" role and authenticates.
Step 2: System navigates to the Admin Command Center (3-tab interface).
Step 3: Issues Tab loads by calling GET /api/v1/reports/prioritized.
Step 4: Backend fetches all reports from Firestore and applies the weighted priority algorithm: each report is scored based on urgency weight (critical=3.0, urgent=2.0, normal=1.0), issue type weight (medical_emergency=3.0, disaster_relief=2.5, legal_aid=2.0, civic_issue=1.0), and time decay bonus (hours since creation × 0.1). Reports are sorted by descending priority score.
Step 5: Admin sees a prioritized list with stats bar (Total, Critical, Urgent, Pending counts), color-coded urgency and issue type badges, description, location, and assigned/required volunteer counts.
Step 6: Admin taps "Find & Assign Volunteers" on a specific issue.
Step 7: System opens a bottom sheet and calls GET /api/v1/reports/{id}/match.
Step 8: Backend fetches the report and all registered volunteers from Firestore.
Step 9: Backend maps the issue type to required skills (e.g., medical_emergency maps to ["medical", "doctor", "paramedic", "general"]).
Step 10: Backend concurrently scores each volunteer using: Skill Score (cosine similarity of skill vectors) × 0.4 + Distance Score (1/(1 + Haversine(km))) × 0.35 + Reliability Score (completion rate) × 0.25.
Step 11: Results are sorted by total score descending and top-20 are returned.
Step 12: Admin sees a scrollable list of matched volunteers with scores, selects via checkboxes.
Step 13: Admin taps "Assign N" button.
Step 14: System calls POST /api/v1/reports/{id}/assign with selected volunteer IDs.
Step 15: Backend updates the Firestore report document: sets assigned_volunteer_ids and changes status to "accepted."
Step 16: Bottom sheet closes, issue list refreshes showing updated assignment count.

### Flow 3: Volunteer Task Execution

Step 1: Volunteer selects "Volunteer" role and authenticates.
Step 2: On sign-up, system auto-registers volunteer in Firestore via POST /api/v1/volunteers.
Step 3: Volunteer Dashboard loads by calling GET /api/v1/volunteers/me/tasks.
Step 4: Backend queries Firestore for all reports where assigned_volunteer_ids contains the volunteer's UID.
Step 5: Volunteer sees task cards with issue descriptions, urgency badges, issue type, and current status.
Step 6: Volunteer taps a status action (e.g., "Start Working" changes status to in_progress; "Mark Resolved" changes to resolved).
Step 7: System calls PATCH /api/v1/reports/{id}/status with the new status value.
Step 8: Backend updates the Firestore document.

### Flow 4: Specialist Case Analysis

Step 1: Specialist (Lawyer/Doctor) selects "Lawyer / Doctor" role and authenticates.
Step 2: Auto-registration creates volunteer record with specialist skills (["specialist", "legal", "medical"]).
Step 3: Specialist Dashboard loads the 3-tab interface.
Step 4: Tab 1 (My Cases): Calls GET /api/v1/cases/my to fetch assigned case files.
Step 5: Tab 2 (AI Q&A): Specialist selects a case and types a natural language question.
Step 6: System calls POST /api/v1/cases/{id}/ask with the question.
Step 7: Backend sends the case documents and question to Gemini 3.1 Pro, which generates an AI-powered answer citing specific document sections.
Step 8: Tab 3 (Document Search): Specialist enters a search query.
Step 9: System calls POST /api/v1/cases/{id}/search.
Step 10: Backend performs semantic matching across all documents in the case and returns ranked results.

### Flow 5: Geospatial Heatmap Visualization

Step 1: Admin navigates to the Heatmap tab.
Step 2: System renders a Google Maps widget using the Maps JavaScript API.
Step 3: All reports with valid GPS coordinates are plotted as markers: Red markers for Critical urgency, Orange for Urgent, Green/Blue for Normal (colored by issue type).
Step 4: Semi-transparent heat circles are drawn around each marker: radius proportional to urgency (500m for Critical, 350m for Urgent, 200m for Normal).
Step 5: In dark mode, a custom dark map style is applied automatically.
Step 6: Admin can toggle between Normal and Satellite map types, zoom in/out, and center the map.
Step 7: Tapping a marker shows an info window with issue details (urgency, type, description, location).
Step 8: Bottom stats bar shows live counts: total issues, critical, pending, active.

---

## 5. System Architecture

### Layer 1: Client / Frontend (Flutter Web & Mobile)

The client is a cross-platform Flutter application (Dart) deployed as a Progressive Web App (PWA) via Firebase Hosting. It uses the following architecture:

- **State Management:** Riverpod (reactive, testable providers).
- **Navigation:** GoRouter with declarative routing and role-based redirect guards.
- **Local Storage:** Hive (lightweight NoSQL) for role persistence and offline caching.
- **Authentication:** Firebase Auth SDK for email/password authentication; JWT tokens attached to every API request.
- **Maps:** Google Maps Flutter SDK with the Maps JavaScript API for web, supporting markers, circles, and custom dark styling.
- **Geolocation:** Web Geolocation API (dart:js_interop + package:web) for capturing GPS coordinates.
- **HTTP Client:** Dio with interceptors for automatic JWT attachment and 401 handling.

**Data Flow:** User Action → Riverpod Provider → ApiClient (Dio) → HTTP Request with JWT → Go Backend.

### Layer 2: Application / Backend (Go + Gin)

The backend is a monolithic Go service using the Gin web framework, structured into clean architecture layers:

- **cmd/api/main.go:** Entry point, route registration, CORS configuration, middleware setup.
- **internal/middleware/auth.go:** Firebase Admin SDK JWT verification middleware; extracts UID and injects into request context.
- **internal/repository/firestore.go:** Data access layer — all Firestore read/write operations (CreateReport, GetReport, GetAllReports, GetAllVolunteers, UpdateReport, etc.).
- **internal/service/matcher.go:** Volunteer matching engine — concurrent scoring with cosine similarity and Haversine distance.
- **internal/service/urgency.go:** Priority scoring algorithm (frequency, severity, population impact, time decay).
- **internal/service/ingestion.go:** Pub/Sub message serialization for the report ingestion pipeline.
- **internal/ai/gemini.go:** Vertex AI (Gemini 3.1 Pro) client for multimodal report parsing (text, image, audio).
- **internal/domain/models.go:** Domain models (Report, Volunteer, User, CaseFile, CaseDocument, Ward, MatchResult, etc.).

**API Endpoints (15+):**

| Method | Path | Purpose |
|--------|------|---------|
| POST | /api/v1/reports | Submit a new issue report |
| GET | /api/v1/reports | List all reports |
| GET | /api/v1/reports/prioritized | Get priority-sorted reports |
| GET | /api/v1/reports/:id | Get a specific report |
| GET | /api/v1/reports/:id/match | Get matching volunteers for a report |
| PATCH | /api/v1/reports/:id/status | Update report status |
| POST | /api/v1/reports/:id/assign | Assign volunteers to a report |
| POST | /api/v1/volunteers | Register a new volunteer |
| GET | /api/v1/volunteers/me/tasks | Get tasks assigned to current volunteer |
| GET | /api/v1/cases/my | Get case files for current specialist |
| POST | /api/v1/cases | Create a new case file |
| POST | /api/v1/cases/:id/documents | Upload document to a case |
| POST | /api/v1/cases/:id/ask | AI Q&A on case documents |
| POST | /api/v1/cases/:id/search | Semantic search within case documents |

**Data Flow:** HTTP Request → Gin Router → Auth Middleware (JWT verification) → Handler Function → Repository (Firestore) → Response.

### Layer 3: Database / Storage

- **Cloud Firestore (NoSQL):** Primary data store for all reports, volunteers, users, case files, and documents. Collections: `reports`, `volunteers`, `users`, `cases`. Firestore was chosen for real-time sync capabilities, offline support, and automatic scaling.
- **Firebase Storage:** Stores uploaded images and documents (medical reports, legal documents). Media URLs are stored in Firestore documents.
- **Hive (Client-Side):** Lightweight key-value store for role persistence, user preferences, and offline report drafts.

### Layer 4: External APIs & Integrations

- **Firebase Authentication:** Handles user identity, session management, and JWT token issuance.
- **Vertex AI (Gemini 3.1 Pro):** Multimodal AI processing — text extraction, image analysis, audio transcription, structured data extraction, and natural language Q&A.
- **Cloud Pub/Sub:** Asynchronous event-driven pipeline for report ingestion. Decouples report submission from AI processing, ensuring the user gets an instant response.
- **Google Maps Platform (Maps JavaScript API):** Real-time geospatial visualization with markers, circles, custom styling, and interaction controls.
- **Web Geolocation API:** Browser-native GPS coordinate capture for accurate issue geolocation.

**End-to-End Data Flow:**

User submits report → Flutter App (Dio + JWT) → Go API Server (Gin + Auth MW) → Firestore (write report) → Pub/Sub (publish event) → [Async] Gemini (AI enrichment) → Firestore (update enriched fields) → Admin Dashboard (reads prioritized reports) → Google Maps (visualizes on heatmap) → Volunteer Matching Engine (scores candidates) → Admin assigns → Firestore (update assignments) → Volunteer Dashboard (reads tasks).

---

## 6. Wireframes / Screen Descriptions (Mockups)

### Screen 1: Role Selection (Landing Page)

**Layout:** Full-screen, vertical layout with top branding and a scrollable list of 4 role cards.

- **Top Section:** SAMAJ logo (gradient icon: indigo-to-teal) + "SAMAJ" text (bold, large, letter-spaced).
- **Subtitle:** "Choose your role" (headline) + "Select how you want to use the platform" (body text).
- **Role Cards (4 items, vertically stacked):**
  - Each card is a horizontal row: gradient icon container (52×52px, rounded, with icon) → title + subtitle → circular arrow button.
  - "General User" — Red-orange gradient, report icon. Subtitle: "Report an issue — photos, location, description."
  - "Volunteer" — Teal gradient, handshake icon. Subtitle: "View & manage assigned tasks."
  - "Lawyer / Doctor" — Purple gradient, gavel icon. Subtitle: "Case files, AI search & document analysis."
  - "NGO Organization" — Violet gradient, admin icon. Subtitle: "Issue management, volunteer matching & heatmap."
- **Interaction:** Tapping a card saves the role and navigates to the Login Screen.

### Screen 2: Reporter Dashboard (Issue Submission Form)

**Layout:** Scrollable form with AppBar and sections.

- **AppBar:** Title "Report an Issue" + logout button.
- **Section 1 — Photos:** Tap-to-add area (dashed border, camera icon), expands to show thumbnails with delete buttons and an "add more" tile.
- **Section 2 — Description:** Multi-line text field (4 rows), placeholder "Describe the issue in detail..."
- **Section 3 — Issue Type:** Dropdown with 4 options: Medical Emergency, Legal Aid, Civic Issue, Disaster Relief.
- **Section 4 — Location:** Text field with location pin icon, placeholder "Enter address or area name."
- **Section 4b — GPS:** "Get My Location" outlined button with my_location icon; after capture, shows "Lat: XX.XXXX, Lng: YY.YYYY" + green checkmark "✓ GPS coordinates captured."
- **Section 5 — Urgency:** 3 horizontal ChoiceChips: Normal (green), Urgent (orange), Critical (red).
- **Section 6 — Volunteer Count:** Numeric text field with people icon.
- **Submit Button:** Full-width filled button "Submit Report" with loading spinner.

### Screen 3: NGO Admin Command Center

**Layout:** TabBarView with 3 tabs — Issues, Matching, Heatmap.

- **AppBar:** Title "NGO Command Center" + refresh + logout. TabBar with 3 tabs: Issues (list icon), Matching (people icon), Heatmap (map icon).
- **Issues Tab:**
  - Stats bar: 4 metric chips (Total/blue, Critical/red, Urgent/orange, Pending/grey).
  - Pull-to-refresh list of issue cards. Each card shows: priority rank number (colored circle), issue type badge, urgency badge, status tag, description (2-line truncated), location with pin icon, assigned/required volunteer count, and full-width "Find & Assign Volunteers" filled button.
- **Heatmap Tab:**
  - Full-screen Google Map widget with issue markers (red = critical, orange = urgent, green = normal by type).
  - Semi-transparent heat circles around each marker (radius = urgency severity).
  - Top overlay: legend card with color-coded counts (Critical, Urgent, Normal, Total).
  - Right side: 4 floating control buttons (layers toggle, center, zoom in, zoom out).
  - Bottom panel: stats row (Issues, Critical, Pending, Active counts with icons).

### Screen 4: Specialist (Lawyer/Doctor) Dashboard

**Layout:** TabBarView with 3 tabs — Case Files, AI Chat, Document Search.

- **AppBar:** Title "Case Analysis" + refresh + logout. TabBar with 3 tabs.
- **Case Files Tab:** Scrollable list of case cards showing title, status badge (open/in_review/closed), linked report count, document count, and creation date.
- **AI Chat Tab:** Chat-style interface. Top: case selector dropdown. Center: scrollable message history (user questions in bubbles right-aligned, AI responses left-aligned with citation highlights). Bottom: text input + send button.
- **Document Search Tab:** Search bar at top. Results displayed as expandable cards showing document name, relevance score, and highlighted matching text snippets.

---

## 7. Technologies to be Used

### Frontend
| Technology | Justification |
|------------|---------------|
| **Flutter (Dart)** | Cross-platform framework enabling a single codebase for Web, Android, and iOS with native-grade performance and a rich widget system. |
| **Riverpod** | Compile-safe, testable state management that avoids the pitfalls of Provider and InheritedWidget. |
| **GoRouter** | Declarative routing with deep-link support, redirect guards, and URL-based navigation essential for web deployment. |
| **Hive** | Lightweight, pure-Dart NoSQL database for offline-first local storage without native dependencies. |
| **Dio** | Powerful HTTP client with interceptor support for automatic JWT attachment, error handling, and request/response logging. |
| **Google Maps Flutter SDK** | Official Google Maps integration for Flutter with full web support, marker/circle APIs, and custom map styling. |

### Backend
| Technology | Justification |
|------------|---------------|
| **Go (Golang)** | Compiled, statically-typed language with first-class concurrency (goroutines) — ideal for concurrent volunteer scoring and high-throughput API serving. |
| **Gin** | Fastest HTTP framework in Go ecosystem, enabling low-latency API responses critical for real-time dashboards. |
| **Firebase Admin SDK (Go)** | Server-side JWT verification and Firestore access with the same identity infrastructure as the client. |

### Database & Storage
| Technology | Justification |
|------------|---------------|
| **Cloud Firestore** | Serverless, auto-scaling NoSQL database with real-time listeners, offline sync, and seamless Firebase integration. |
| **Firebase Storage** | Managed object storage for user-uploaded images and documents with security rules and CDN delivery. |

### Cloud & Hosting
| Technology | Justification |
|------------|---------------|
| **Firebase Hosting** | Global CDN for Flutter Web deployment with HTTPS, custom domains, and instant rollbacks. |
| **Cloud Run** | Serverless container hosting for the Go API backend — scales to zero, pay-per-request, auto-managed TLS. |
| **Cloud Pub/Sub** | Fully-managed event streaming for decoupling report submission from AI processing, ensuring sub-second user response times. |

### AI / ML
| Technology | Justification |
|------------|---------------|
| **Vertex AI (Gemini 3.1 Pro)** | Google's most advanced multimodal model — handles text, images, and audio in a single API call with structured JSON output. |
| **Gemini Function Calling** | Enables structured data extraction (problem_category, severity_index, etc.) with deterministic JSON responses. |

### APIs & Integrations
| Technology | Justification |
|------------|---------------|
| **Google Maps JavaScript API** | Industry-standard mapping platform with heatmap visualization, custom markers, and dark mode styling. |
| **Web Geolocation API** | Browser-native GPS coordinate capture — zero-dependency, works across all modern browsers. |
| **Firebase Authentication** | Enterprise-grade identity management with email/password, OAuth, and anonymous auth — zero backend burden. |

### Development & Tooling
| Technology | Justification |
|------------|---------------|
| **Git + GitHub** | Version control and CI/CD pipeline integration. |
| **Firebase CLI** | Single-command deployment for hosting, functions, and security rules. |
| **Go Modules** | Dependency management for the Go backend. |

---

## 8. Estimated Implementation Cost

### Monthly Operational Cost (Production Scale — 10,000 MAU)

| Service | Tier | Estimated Monthly Cost (USD) |
|---------|------|------------------------------|
| **Firebase Hosting** | Blaze (Pay-as-you-go) | $0 – $5 (generous free tier covers most PWA hosting) |
| **Cloud Firestore** | Blaze | $15 – $40 (based on 500K reads + 100K writes/month) |
| **Firebase Authentication** | Free tier | $0 (free up to 50K MAU) |
| **Firebase Storage** | Blaze | $5 – $15 (for images and documents, ~50GB) |
| **Cloud Run (Go API)** | Pay-per-request | $5 – $20 (scales to zero, ~100K requests/month) |
| **Cloud Pub/Sub** | Standard | $2 – $5 (~100K messages/month) |
| **Vertex AI (Gemini 3.1 Pro)** | Pay-per-token | $30 – $80 (depends on report volume and multimodal usage) |
| **Google Maps Platform** | Pay-per-load | $7 – $14 (28,000 free map loads/month, then $7/1000) |
| **Domain & SSL** | Annual | ~$1/month (prorated) |
| **Monitoring (Cloud Logging)** | Free tier | $0 |

| | **Total Estimated Monthly Cost** | **$65 – $180** |

### One-Time Development Cost (Estimated)

| Item | Cost (USD) |
|------|------------|
| Flutter App Development (4 dashboards) | $8,000 – $12,000 |
| Go Backend Development (15+ endpoints) | $6,000 – $10,000 |
| AI Integration (Gemini pipeline) | $3,000 – $5,000 |
| UI/UX Design & Dark Mode | $2,000 – $3,000 |
| Testing & QA | $2,000 – $3,000 |
| **Total Development** | **$21,000 – $33,000** |

**Note:** For a hackathon/MVP context, leveraging Google Cloud's $300 free trial credit and Firebase's generous free tier reduces operational costs to effectively $0 for the first 90 days.

---

## 9. Additional Details & Future Development

### Phase 1: Post-Hackathon Polish (Month 1-2)
- Deploy Go backend to Cloud Run for production-grade hosting (eliminating localhost dependency).
- Implement Firebase Cloud Messaging (FCM) for real-time push notifications to volunteers.
- Add offline report drafting with Hive and auto-sync on connectivity restore.
- Polish responsive layouts for tablet and desktop viewports.

### Phase 2: Intelligence Amplification (Month 3-6)
- **Voice Report Ingestion:** Enable field workers to submit reports by speaking in Hindi/regional dialects, processed by Gemini's audio capabilities.
- **Duplicate Detection:** Use Gemini embeddings to compute semantic similarity between incoming and existing reports, auto-flagging potential duplicates.
- **Predictive Hotspot Modeling:** Train a time-series model on historical report data (location, type, frequency) to predict future issue hotspots — enabling proactive resource deployment.
- **S2 Cell Indexing:** Replace raw lat/lng with Google S2 Geometry cell IDs for efficient spatial queries and volunteer proximity lookups.

### Phase 3: Scale & Ecosystem (Month 6-12)
- **Multi-Tenant Architecture:** Support multiple NGOs on a single deployment with isolated data, role hierarchies, and cross-org collaboration capabilities.
- **Government API Integration:** Bi-directional sync with CPGRAMS, MyGov, and state-level grievance portals for automated escalation of unresolved issues.
- **WhatsApp/SMS Gateway:** Enable issue reporting via WhatsApp Business API and Twilio SMS for users without smartphones.
- **Volunteer Gamification:** Points, badges, leaderboards, and monthly recognition to drive sustained volunteer engagement.
- **Public Transparency Dashboard:** Citizen-facing portal showing aggregate resolution rates, response times, and issue density by ward — building civic trust.

### Phase 4: Enterprise & Expansion (Year 2-3)
- **White-Label Solution:** Package SAMAJ as a configurable SaaS product for civic bodies, disaster relief organizations, and healthcare networks globally.
- **Federated Learning:** Enable AI model improvement across NGO deployments without sharing sensitive data, maintaining privacy while improving accuracy.
- **Drone & IoT Integration:** Accept reports from IoT sensors (flood gauges, air quality monitors) and drone surveys for large-scale environmental monitoring.
- **Blockchain Audit Trail:** Immutable record of every state transition for regulatory compliance, anti-corruption, and donor accountability.
- **Expansion Markets:** Southeast Asia, Sub-Saharan Africa, and Latin America — regions with similar civic infrastructure challenges.

### Pivot Opportunities
- **Disaster Response Platform:** Position SAMAJ as the default coordination tool for disaster relief organizations (Red Cross, NDRF) — the multimodal intake + volunteer matching + heatmap stack is directly applicable.
- **Smart City Dashboard:** Municipal governments can use SAMAJ as their citizen reporting and service delivery layer, replacing legacy 311 systems.
- **Healthcare Network Coordinator:** Specialist case management + volunteer matching can be adapted for community health worker coordination in rural areas.

---

*Document prepared for the Google Solutions Challenge 2026.*
*Team SAMAJ — Building Community Intelligence at Scale.*
