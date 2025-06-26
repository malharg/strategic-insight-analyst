// frontend/app/dashboard/page.tsx
"use client";

import { useState, ChangeEvent, FormEvent, useEffect, useCallback } from "react";
import { useAuth } from "@/context/AuthContext";
import { useRouter } from "next/navigation";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { authenticatedFetch } from "@/lib/api";
import { auth } from "@/lib/firebase";
import Link from "next/link";

// Type definition for a document
type Document = {
  id: string;
  fileName: string;
  uploadedAt: string;
};

export default function DashboardPage() {
  const { user, loading } = useAuth();
  const router = useRouter();

  const [file, setFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [documents, setDocuments] = useState<Document[]>([]);

  const fetchDocuments = useCallback(async () => {
    if (!user) return;
    try {
      const docs = await authenticatedFetch("/api/documents");
      if (Array.isArray(docs)) {
        setDocuments(docs);
      } else {
        setDocuments([]);
      }
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError("Failed to fetch documents: " + err.message);
      } else {
        setError("An unknown error occurred while fetching documents.");
      }
    }
  }, [user]);

  useEffect(() => {
    if (!loading && !user) {
      router.push("/");
    }
    if (user) {
      fetchDocuments();
    }
  }, [user, loading, router, fetchDocuments]);

  const handleDelete = async (docId: string) => {
    setDocuments(currentDocs => currentDocs.filter(doc => doc.id !== docId));
    setError("");

    try {
      await authenticatedFetch(`/api/documents/delete?id=${docId}`, {
        method: "DELETE",
      });
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(`Failed to delete document. Please refresh. Error: ${err.message}`);
      } else {
        setError("An unknown error occurred while deleting the document.");
      }
      fetchDocuments();
    }
  };

  const handleFileChange = (e: ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      setFile(e.target.files[0]);
      setMessage("");
      setError("");
    }
  };

  const handleUpload = async (e: FormEvent) => {
    e.preventDefault();
    if (!file) {
      setError("Please select a file first.");
      return;
    }
    setIsUploading(true);
    setMessage("Uploading...");
    setError("");
    const formData = new FormData();
    formData.append("document", file);
    try {
      await authenticatedFetch("/api/documents/upload", {
        method: "POST",
        body: formData,
      });
      setMessage("Upload successful!");
      setFile(null);
      await fetchDocuments();
      setMessage("");
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("An unexpected error occurred during upload.");
      }
    } finally {
      setIsUploading(false);
    }
  };

  if (loading) {
    return <div className="flex h-screen items-center justify-center">Loading session...</div>;
  }

  if (!user) {
    return null;
  }

  return (
    <div className="flex flex-col min-h-screen bg-gray-50">
      <header className="flex items-center justify-between p-4 bg-white border-b">
        <h1 className="text-xl font-bold">Strategic Insight Analyst</h1>
        <div className="flex items-center gap-4">
          <span className="text-sm text-gray-600">{user.email}</span>
          <Button variant="outline" onClick={() => auth.signOut()}>Sign Out</Button>
        </div>
      </header>

      <main className="flex-grow container mx-auto p-4 md:p-8">
        <Card className="w-full max-w-2xl mx-auto">
          <CardHeader>
            <CardTitle>Upload a New Document</CardTitle>
            <CardDescription>Upload a PDF or TXT file for analysis.</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleUpload} className="space-y-6">
              <div className="grid w-full items-center gap-1.5">
                <Label htmlFor="document">Document File</Label>
                <Input id="document" type="file" onChange={handleFileChange} accept=".pdf,.txt" />
              </div>
              <Button type="submit" disabled={!file || isUploading} className="w-full">
                {isUploading ? "Uploading..." : "Upload and Process"}
              </Button>
              {message && <p className="text-sm font-medium text-green-600">{message}</p>}
              {error && <p className="text-sm font-medium text-destructive">{error}</p>}
            </form>
          </CardContent>
        </Card>

        <Card className="w-full max-w-2xl mx-auto mt-8">
          <CardHeader><CardTitle>My Documents</CardTitle></CardHeader>
          <CardContent>
            {documents.length > 0 ? (
              <ul className="space-y-2">
                {documents.map((doc) => (
                  <li key={doc.id} className="flex justify-between items-center p-3 border rounded hover:bg-gray-50">
                    <div className="flex flex-col flex-1 min-w-0 mr-4">
                      <span className="font-medium truncate">{doc.fileName}</span>
                      <span className="text-xs text-gray-500">
                        Uploaded on: {new Date(doc.uploadedAt).toLocaleDateString()}
                      </span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Link href={`/chat/${doc.id}`} passHref>
                        <Button variant="outline" size="sm">Analyze</Button>
                      </Link>
                      <AlertDialog>
                        <AlertDialogTrigger asChild>
                          <Button variant="destructive" size="sm">Delete</Button>
                        </AlertDialogTrigger>
                        <AlertDialogContent>
                          <AlertDialogHeader>
                            <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
                            <AlertDialogDescription>
                              This action cannot be undone. This will permanently delete the
                              document and all associated analysis history.
                            </AlertDialogDescription>
                          </AlertDialogHeader>
                          <AlertDialogFooter>
                            <AlertDialogCancel>Cancel</AlertDialogCancel>
                            <AlertDialogAction onClick={() => handleDelete(doc.id)}>
                              Continue
                            </AlertDialogAction>
                          </AlertDialogFooter>
                        </AlertDialogContent>
                      </AlertDialog>
                    </div>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-sm text-gray-500">
                You haven&rsquo;t uploaded any documents yet.
              </p>
            )}
          </CardContent>
        </Card>
      </main>
    </div>
  );
}
