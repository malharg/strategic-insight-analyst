# Strategic Insight Analyst

This is a full-stack application built for the SDE Internship Assignment. The application allows users to upload business documents (PDF or TXT) and use an AI model (Google Gemini) to extract, summarize, and analyze strategic insights from them.

**Live Demo:** [Link to your deployed Vercel URL]
**Video Walkthrough:** [Link to your 2-3 minute video]

---

## Key Features

-   **Secure User Authentication:** Email/Password login system using Firebase Authentication. Users can only access their own documents.
-   **Full Document Management:** Upload, list, and delete documents. Original files are stored securely in a cloud object store.
-   **AI-Powered Analysis:** An interactive chat interface allows users to query their documents. The backend constructs sophisticated prompts and uses Google's Gemini LLM to generate insights.
-   **Transactional Database:** All metadata, extracted text chunks, and chat history are stored in a SQL database, ensuring data integrity.
-   **Fully Deployed:** The frontend is deployed on Vercel and the backend on Google Cloud Run, demonstrating a complete, production-ready system.

---

## Technical Stack

### Backend
-   **Language:** Go (GoLang)
-   **API:** Standard Library `net/http`
-   **Authentication:** Firebase Admin SDK (for token verification)
-   **Database:** SQLite (for local development), PostgreSQL (for production via Supabase)
-   **Cloud Storage:** Supabase Storage (backed by S3)
-   **PDF Processing:** UniDoc (`unipdf/v3`)
-   **AI Integration:** Google Gemini (`gemini-1.5-flash`) via REST API
-   **Deployment:** Docker, Google Cloud Run

### Frontend
-   **Framework:** Next.js (App Router)
-   **Language:** TypeScript
-   **UI Library:** ShadCN UI & Tailwind CSS
-   **Authentication:** Firebase Client SDK
-   **State Management:** React Context API & `useState`
-   **Deployment:** Vercel

---

## Setup & Local Development

### Prerequisites
-   [Go](https://go.dev/doc/install) (version 1.21+ recommended)
-   [Node.js](https://nodejs.org/) (version 18+ recommended)
-   [Docker Desktop](https://www.docker.com/products/docker-desktop/) (must be running for local development)
-   [gcloud CLI](https://cloud.google.com/sdk/docs/install) (for deployment)

### API Keys & Environment Variables

You will need the following accounts and API keys to run this project:
-   **Firebase Project:** For user authentication.
-   **Supabase Project:** For cloud storage and the PostgreSQL database.
-   **Google AI Studio:** For the Gemini API key.
-   **UniDoc:** For a free metered license key to process PDFs.

#### Backend `.env` File
Create a file named `.env` in the `/backend` directory and add the following keys:

```
# From your Firebase project settings
FIREBASE_PROJECT_ID="your-firebase-project-id"

# From your Supabase project settings (API section)
SUPABASE_URL="https://<your-ref-id>.supabase.co"
SUPABASE_SERVICE_KEY="your-long-supabase-service-role-key"

# From Google AI Studio (aistudio.google.com)
GEMINI_API_KEY="your-gemini-api-key"

# From UniDoc (unidoc.io/license)
UNIDOC_LICENSE_KEY="your-unidoc-license-key"
```

#### Frontend `.env.local` File
Create a file named `.env.local` in the `/frontend` directory and add the following keys from your Firebase project's web app configuration:

```
NEXT_PUBLIC_FIREBASE_API_KEY="AIza..."
NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN="..."
NEXT_PUBLIC_FIREBASE_PROJECT_ID="..."
NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET="..."
NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID="..."
NEXT_PUBLIC_FIREBASE_APP_ID="..."
```

### Running Locally

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/malharg/strategic-insight-analyst.git
    cd strategic-insight-analyst
    ```

2.  **Start the Backend Server:**
    -   Navigate to the backend directory: `cd backend`
    -   Install dependencies: `go mod tidy`
    -   Run the server: `go run main.go`
    -   The backend will be running on `http://localhost:8080`.

3.  **Start the Frontend Server:**
    -   In a new terminal, navigate to the frontend directory: `cd frontend`
    -   Install dependencies: `npm install`
    -   Run the development server: `npm run dev`
    -   The frontend will be running on `http://localhost:3000`.

---

## Deployment

### Backend (Google Cloud Run)
The backend is containerized with Docker and deployed to Google Cloud Run.

1.  **Authenticate with gcloud:**
    ```bash
    gcloud auth login
    gcloud config set project <your-gcp-project-id>
    ```
2.  **Deploy from the `/backend` directory:**
    ```bash
    gcloud run deploy strategic-insight-backend --source . --region <your-region> --allow-unauthenticated
    ```
3.  After deployment, set the environment variables in the Cloud Run service console using the values from your `.env` file.

### Frontend (Vercel)
1.  Push the latest code to the GitHub repository.
2.  Import the project into Vercel from your GitHub account.
3.  Configure the environment variables in the Vercel project settings using the values from your `.env.local` file.
4.  **Crucially**, add the `NEXT_PUBLIC_API_BASE_URL` variable and set its value to the public URL of your deployed Cloud Run service.
5.  Click "Deploy".