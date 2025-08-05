import { clerkMiddleware, createRouteMatcher } from '@clerk/nextjs/server'
import { NextResponse } from 'next/server'

const isPublicRoute = createRouteMatcher([
  '/',
  '/sign-in(.*)',
  '/sign-up(.*)',
  '/api(.*)',  // API routes should be accessible
])

export default clerkMiddleware(async (auth, req) => {
  const { userId, sessionClaims, redirectToSignIn } = await auth()

  // For public routes, do nothing
  if (isPublicRoute(req)) {
    return NextResponse.next()
  }

  // If the user isn't authenticated, redirect to sign-in
  if (!userId) {
    return redirectToSignIn()
  }

  // Check user status from session claims
  const status = (sessionClaims as any)?.status as string | undefined
  const userRole = (sessionClaims as any)?.user_role as string | undefined
  
  // If user is pending and not on the setup page, redirect them
  if (status === 'pending' && req.nextUrl.pathname !== '/profile/setup') {
    const setupUrl = new URL('/profile/setup', req.url)
    return NextResponse.redirect(setupUrl)
  }

  // If user is active but tries to access setup, redirect to spaces
  if (status === 'active' && req.nextUrl.pathname === '/profile/setup') {
    const spacesUrl = new URL('/spaces', req.url)
    return NextResponse.redirect(spacesUrl)
  }

  // If no status (new user), redirect to profile setup for protected routes
  if (!status && req.nextUrl.pathname !== '/profile/setup') {
    const setupUrl = new URL('/profile/setup', req.url)
    return NextResponse.redirect(setupUrl)
  }

  return NextResponse.next()
})

export const config = {
  matcher: [
    // Skip Next.js internals and all static files, unless found in search params
    '/((?!_next|[^?]*\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)',
    // Always run for API routes
    '/(api|trpc)(.*)',
  ],
}
