import { Outlet } from "@tanstack/react-router";
import Header from "./Header";
import Footer from "./Footer";
import { useAuthRefresh } from "@/hook";

export default function RootLayout() {
  useAuthRefresh();

  return (
    <div className="flex flex-col min-h-screen bg-gray-50">
      <Header />
      <main className="flex-1 w-full py-6 sm:py-8">
        <Outlet />
      </main>
      <Footer />
    </div>
  );
}
