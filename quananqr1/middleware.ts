import { Role, RoleType } from "@/constants/type";
import { decodeToken } from "@/lib/utils";
import { NextRequest, NextResponse } from "next/server";

// Define path configurations
const adminPaths = ["/manage/admin"];
const employeePaths = ["/manage/employee"];

const privatePaths = [...adminPaths, ...employeePaths];
const unAuthPaths = ["/auth"];
const wellComePage = ["/"];

// Define allowed roles for different paths
const pathRoleConfig: Record<string, RoleType[]> = {
  "/manage/admin": [Role.Admin],
  "/manage/employee": [Role.Employee, Role.Admin]
};

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const accessToken = request.cookies.get("accessToken")?.value;
  if (pathname === "/manage" || pathname.startsWith("/manage/")) {
    if (!accessToken) {
      const url = new URL("/auth", request.url);
      url.searchParams.set("from", pathname);
      return NextResponse.redirect(url);
    }
    try {
      const decoded = decodeToken(accessToken);
      console.log("quananqr1/middleware.ts decoded", decoded);
      const userRole = decoded.role as RoleType;
      // Check for specific manage routes
      if (pathname === "/manage") {
        // Redirect to appropriate dashboard based on role
        if (userRole === Role.Admin) {
          return NextResponse.redirect(new URL("/manage/admin", request.url));
        } else if (userRole === Role.Employee) {
          return NextResponse.redirect(
            new URL("/manage/employee", request.url)
          );
        }
      }
      // Check role-based access for specific manage routes
      for (const [path, allowedRoles] of Object.entries(pathRoleConfig)) {
        if (pathname.startsWith(path) && !allowedRoles.includes(userRole)) {
          // Redirect to welcome page if not authorized
          return NextResponse.redirect(new URL("/", request.url));
        }
      }
    } catch (error) {
      // Token decode failed - redirect to login
      const url = new URL("/auth", request.url);
      url.searchParams.set("from", pathname);
      return NextResponse.redirect(url);
    }
  }
  // Previous authentication and redirect logic remains the same
  if (unAuthPaths.includes(pathname) && accessToken) {
    const fromPath = request.nextUrl.searchParams.get("from");
    if (fromPath && fromPath !== "/") {
      return NextResponse.redirect(new URL(fromPath, request.url));
    }
    return NextResponse.redirect(new URL("/", request.url));
  }
  if (privatePaths.some((path) => pathname.startsWith(path)) && !accessToken) {
    const url = new URL("/auth", request.url);
    url.searchParams.set("from", pathname);
    return NextResponse.redirect(url);
  }
  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!api|_next/static|_next/image|favicon.ico).*)"]
};
