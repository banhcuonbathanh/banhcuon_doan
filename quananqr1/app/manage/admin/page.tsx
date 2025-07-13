import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { decodeToken } from "@/lib/utils";
import { Role } from "@/constants/type";
import WebSocketWrapper from "@/components/websocket/WebSocketWrapper";


export default async function ManageHomePage() {
  console.log("quananqr1/app/manage/admin/page.tsx ManageHomePage");
  const cookieStore = cookies();
  const accessToken = cookieStore.get("accessToken")?.value;

  // Double-check authorization on server side
  if (!accessToken) {
    redirect("/login");
  }

  try {
    const decoded = decodeToken(accessToken);
    if (!(decoded.role === Role.Admin || decoded.role === Role.Employee)) {
      redirect("/manage/employee");
    }

    console.log("quananqr1/app/manage/admin/page.tsx ManageHomePage 333");
    return (
      <div className="p-4">
        <h1 className="text-2xl font-bold">Manage Dashboard</h1>
        <WebSocketWrapper accessToken={accessToken} />
      </div>
    );
  } catch (error) {
    redirect("/auth");
  }
}