import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

export default function HomePage() {
  return (
    <div className="flex min-h-screen flex-col">
      {/* Header */}
      <header className="border-b bg-white">
        <div className="container mx-auto flex h-16 items-center justify-between px-4">
          <h1 className="text-2xl font-bold">LibSystem</h1>
          <div className="flex gap-4">
            <Link href="/login">
              <Button variant="outline">Sign in</Button>
            </Link>
            <Link href="/register">
              <Button>Get started</Button>
            </Link>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <main className="flex-1">
        <section className="container mx-auto px-4 py-12 sm:py-20">
          <div className="mx-auto max-w-3xl text-center">
            <h2 className="text-3xl font-bold tracking-tight sm:text-4xl lg:text-6xl">
              Digital Library Management Made Simple
            </h2>
            <p className="mt-6 text-base sm:text-lg leading-7 sm:leading-8 text-gray-600">
              Organize, search, and access your digital documents with powerful tools.
              Upload files, create collections, and find what you need instantly.
            </p>
            <div className="mt-8 sm:mt-10 flex flex-col sm:flex-row items-center justify-center gap-4 sm:gap-6">
              <Link href="/register" className="w-full sm:w-auto">
                <Button size="lg" className="w-full sm:w-auto">Get started</Button>
              </Link>
              <Link href="/login" className="w-full sm:w-auto">
                <Button variant="outline" size="lg" className="w-full sm:w-auto">
                  Sign in
                </Button>
              </Link>
            </div>
          </div>
        </section>

        {/* Features Section */}
        <section className="border-t bg-gray-50 py-12 sm:py-20">
          <div className="container mx-auto px-4">
            <h3 className="text-center text-2xl sm:text-3xl font-bold">Key Features</h3>
            <div className="mt-16 grid gap-8 sm:grid-cols-2 lg:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle>Document Upload</CardTitle>
                  <CardDescription>
                    Upload PDFs, Word documents, images, and more
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-gray-600">
                    Support for multiple file formats including PDF, DOCX, TXT, images, and HTML files up to 100MB.
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Full-Text Search</CardTitle>
                  <CardDescription>
                    Find documents instantly with powerful search
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-gray-600">
                    Search across all your documents with faceted filters, highlighting, and relevance ranking.
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Collections</CardTitle>
                  <CardDescription>
                    Organize documents into collections
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-gray-600">
                    Create public or private collections to organize your documents and share with others.
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>OCR Support</CardTitle>
                  <CardDescription>
                    Extract text from images automatically
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-gray-600">
                    Automatic optical character recognition for scanned documents and images.
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Role-Based Access</CardTitle>
                  <CardDescription>
                    Control who can access what
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-gray-600">
                    Admin, librarian, and patron roles with appropriate permissions for each.
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Real-Time Indexing</CardTitle>
                  <CardDescription>
                    Documents searchable immediately
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-gray-600">
                    Automatic background indexing makes your documents searchable within seconds of upload.
                  </p>
                </CardContent>
              </Card>
            </div>
          </div>
        </section>
      </main>

      {/* Footer */}
      <footer className="border-t bg-white py-8">
        <div className="container mx-auto px-4 text-center text-sm text-gray-600">
          <p>&copy; 2026 LibSystem. All rights reserved.</p>
        </div>
      </footer>
    </div>
  );
}
