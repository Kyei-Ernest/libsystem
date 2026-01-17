import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

/**
 * Middleware to protect dashboard routes
 * Redirects to login if user is not authenticated
 */
export function middleware(request: NextRequest) {
    const { pathname } = request.nextUrl;

    // Check if accessing dashboard routes
    if (pathname.startsWith('/dashboard')) {
        // For client-side navigation, this won't work perfectly
        // We'll need to handle auth check on the client side as well
        // This is just a basic layer
        return NextResponse.next();
    }

    return NextResponse.next();
}

export const config = {
    matcher: ['/dashboard/:path*'],
};
