# LibSystem Frontend

Modern web application for the LibSystem digital library management system.

## Tech Stack

- **Framework:** Next.js 16.1.1 (App Router)
- **Language:** TypeScript 5.9.3
- **Styling:** Tailwind CSS 4.1.18
- **UI Components:** shadcn/ui
- **State Management:** Zustand
- **Data Fetching:** Axios + React Query
- **Forms:** React Hook Form + Zod
- **Authentication:** Custom JWT with localStorage
- **Icons:** Lucide React

## Quick Start

### Prerequisites

- Node.js 18+ (20+ recommended for production builds)
- npm or pnpm
- Backend services running on localhost:8088

### Installation

```bash
# Install dependencies
npm install

# Start development server
npm run dev
```

Visit [http://localhost:3000](http://localhost:3000)

## Project Structure

```
frontend/
├── app/                    # Next.js App Router
│   ├── (auth)/            # Authentication pages
│   │   ├── login/
│   │   └── register/
│   ├── (dashboard)/       # Protected dashboard routes
│   │   ├── documents/
│   │   ├── collections/
│   │   ├── search/
│   │   └── settings/
│   ├── layout.tsx         # Root layout
│   ├── page.tsx           # Landing page
│   └── globals.css        # Global styles
├── components/
│   ├── ui/                # shadcn/ui components
│   ├── features/          # Feature-specific components
│   └── layouts/           # Layout components
├── lib/
│   ├── api/               # API client & modules
│   │   ├── client.ts      # Axios configuration
│   │   └── auth.ts        # Auth endpoints
│   ├── hooks/             # Custom React hooks
│   ├── utils/             # Utility functions
│   └── validations/       # Zod schemas
├── stores/
│   └── auth-store.ts      # Zustand auth state
├── types/
│   └── api.ts             # TypeScript type definitions
├── public/                # Static assets
├── .env.local             # Environment variables
├── tailwind.config.js
├── tsconfig.json
└── next.config.ts
```

## Environment Variables

Create `.env.local` file:

```env
NEXT_PUBLIC_API_URL=http://localhost:8088/api/v1
NEXTAUTH_SECRET=your-secret-key-minimum-32-characters
NEXTAUTH_URL=http://localhost:3000
```

## Available Scripts

```bash
# Development
npm run dev              # Start dev server (with Turbopack)

# Building
npm run build            # Build for production
npm start                # Start production server

# Code Quality
npm run lint             # Run ESLint
npm run type-check       # TypeScript type checking
```

## Features

### Implemented ✅
- User authentication (login/register)
- JWT token management
- Protected routes
- Landing page
- Basic dashboard
- Toast notifications
- Form validation
- Responsive design
- Dark mode support (via Tailwind)

### Planned 🚧
- Document upload with progress tracking
- Document list/grid views with pagination  
- Full-text search with facets
- Collection management (CRUD)
- Document viewer (PDF.js)
- User profile management
- Admin analytics dashboard
- Role-based access control
- Error boundaries
- E2E tests (Playwright)

## API Integration

The frontend connects to the LibSystem backend API Gateway:

**Base URL:** `http://localhost:8088/api/v1`

**Endpoints:**
- `POST /auth/register` - Register new user
- `POST /auth/login` - User login
- `GET /auth/profile` - Get user profile
- `GET /documents` - List documents
- `POST /documents` - Upload document
- `GET /collections` - List collections
- `POST /search` - Search documents

See [`lib/api/`](./lib/api) for implementation details.

## Type Safety

All API responses and data models are strongly typed in [`types/api.ts`](./types/api.ts) to match the backend schemas exactly.

## Development Guidelines

### Code Style
- Use TypeScript for all files
- Follow ESLint configuration
- Use functional components with hooks
- Use server components where possible (Next.js App Router)

### State Management
- **Server state:** React Query (planned)
- **Client state:** Zustand stores
- **Form state:** React Hook Form

### Styling
- Use Tailwind CSS utility classes
- Follow shadcn/ui component patterns
- Mobile-first responsive design

## Testing

```bash
# Unit tests (planned)
npm run test

# E2E tests (planned)
npm run test:e2e
```

## Deployment

### Production Build

```bash
npm run build
npm start
```

### Environment Setup

Ensure backend services are accessible and update `.env.local` accordingly.

## Troubleshooting

**Dev server won't start:**
- Check Node.js version: `node --version` (18+ required)
- Delete `node_modules` and reinstall: `rm -rf node_modules && npm install`
- Clear Next.js cache: `rm -rf .next`

**API requests failing:**
- Verify backend is running: curl http://localhost:8088/api/v1/health
- Check CORS configuration on backend
- Verify `NEXT_PUBLIC_API_URL` in `.env.local`

**Authentication not working:**
- Check localStorage in DevTools
- Verify JWT token format
- Check backend auth service logs

## Contributing

1. Create feature branch
2. Make changes
3. Run `npm run lint`
4. Test thoroughly
5. Submit pull request

## License

MIT

---

**Backend Repository:** `/home/ernest-kyei/Documents/libsystem`  
**Documentation:** See `/docs` directory in main repo  
**API Docs:** http://localhost:8088/api/v1/docs (when running)
