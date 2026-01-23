import { useEffect } from "react";
import { useNavigate } from "@tanstack/react-router";
import { GoogleLogin } from "../components/auth/GoogleLogin";
import { TestLogin } from "../components/auth/TestLogin";
import { useAuth } from "../hook";
import { FileText, Bookmark, ThumbsUp } from "lucide-react";

export default function LoginPage() {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();

  useEffect(() => {
    if (isAuthenticated) {
      navigate({ to: "/feed" });
    }
  }, [isAuthenticated, navigate]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <div className="w-full max-w-md rounded-lg bg-white p-8 shadow-md">
        <div className="mb-8 text-center">
          <h1 className="mb-2 text-3xl font-bold text-gray-900">OpenGov</h1>
          <p className="text-gray-600">
            Stay informed about what your government is doing
          </p>
        </div>

        <div className="mb-8 space-y-3">
          <GoogleLogin />
          <TestLogin />
        </div>

        <div className="space-y-4 mb-8">
          <div className="flex items-start gap-3 text-sm text-gray-600">
            <FileText className="w-5 h-5 text-blue-600 mt-0.5 flex-shrink-0" />
            <p>Access real-time Federal Register updates as they happen</p>
          </div>
          <div className="flex items-start gap-3 text-sm text-gray-600">
            <Bookmark className="w-5 h-5 text-blue-600 mt-0.5 flex-shrink-0" />
            <p>Bookmark important documents for quick access later</p>
          </div>
          <div className="flex items-start gap-3 text-sm text-gray-600">
            <ThumbsUp className="w-5 h-5 text-blue-600 mt-0.5 flex-shrink-0" />
            <p>Track which issues matter most to you</p>
          </div>
        </div>

        <div className="text-center">
          <p className="text-sm text-gray-500">
            By signing in, you agree to our Terms of Service and Privacy Policy
          </p>
        </div>
      </div>
    </div>
  );
}
