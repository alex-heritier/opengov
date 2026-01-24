import { useParams, Link } from "@tanstack/react-router";
import {
  ArrowLeft,
  ExternalLink,
  Calendar,
  Clock,
  AlertCircle,
  Building2,
  Zap,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Alert, AlertDescription } from "@/components/ui/alert";
import ShareButtons from "@/components/share/ShareButtons";
import { useArticleBySlugQuery } from "@/hook";

export default function ArticleDetailPage() {
  const { slug } = useParams({ from: "/articles/$slug" });
  const {
    data: article,
    isLoading: loading,
    error,
  } = useArticleBySlugQuery(slug);

  if (loading) {
    return (
      <div className="w-full max-w-4xl mx-auto px-4 sm:px-6 py-6 sm:py-8 space-y-4">
        <Skeleton className="h-8 w-3/4" />
        <Skeleton className="h-4 w-1/4" />
        <Skeleton className="h-32 w-full" />
      </div>
    );
  }

  if (error || !article) {
    return (
      <div className="w-full max-w-4xl mx-auto px-4 sm:px-6 py-6 sm:py-8">
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            {error?.message ?? "Article not found"}
          </AlertDescription>
        </Alert>
        <Button asChild variant="outline" className="mt-4">
          <Link to="/feed" className="inline-flex items-center gap-2">
            <ArrowLeft className="w-4 h-4" />
            Back to Feed
          </Link>
        </Button>
      </div>
    );
  }

  const formattedPublishedDate = new Date(
    article.published_at,
  ).toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });

  const formattedTime = new Date(article.published_at).toLocaleTimeString(
    "en-US",
    {
      hour: "2-digit",
      minute: "2-digit",
    },
  );

  const hasPoliticalScore =
    article.political_score !== null && article.political_score !== undefined;

  return (
    <div className="w-full max-w-4xl mx-auto px-4 sm:px-6 py-6 sm:py-8">
      {/* Back Button */}
      <Button
        asChild
        variant="ghost"
        className="mb-4 sm:mb-6 text-sm sm:text-base"
      >
        <Link to="/feed" className="inline-flex items-center gap-2">
          <ArrowLeft className="w-4 h-4" />
          Back to Feed
        </Link>
      </Button>

      {/* Article Header */}
      <article className="bg-white rounded-lg border border-gray-200 overflow-hidden">
        {/* Title Banner */}
        <div className="bg-gray-50 border-b border-gray-200 px-4 sm:px-8 py-4 sm:py-6">
          <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 mb-3">
            {article.title}
          </h1>
          <div className="flex flex-wrap gap-3 sm:gap-4 text-gray-600 text-xs sm:text-sm">
            {article.agency && (
              <div className="flex items-center gap-2">
                <Building2 className="w-4 h-4" />
                <span>{article.agency}</span>
              </div>
            )}
            <div className="flex items-center gap-2">
              <Calendar className="w-4 h-4" />
              <span>{formattedPublishedDate}</span>
            </div>
            <div className="flex items-center gap-2">
              <Clock className="w-4 h-4" />
              <span>{formattedTime}</span>
            </div>
            {article.document_number && (
              <div className="flex items-center gap-2">
                <span className="font-semibold">Doc #:</span>
                <span className="font-mono text-xs">
                  {article.document_number}
                </span>
              </div>
            )}
          </div>
        </div>

        {/* Article Content */}
        <div className="px-4 sm:px-8 py-4 sm:py-6 space-y-4 sm:space-y-6">
          {/* Summary */}
          <div>
            <h2 className="text-lg sm:text-xl font-bold text-gray-900 mb-3">
              Summary
            </h2>
            <p className="text-sm sm:text-base text-gray-700 leading-relaxed whitespace-pre-wrap">
              {article.summary}
            </p>
          </div>

          {/* Key Points */}
          {article.keypoints && article.keypoints.length > 0 && (
            <div>
              <h2 className="text-lg sm:text-xl font-bold text-gray-900 mb-3">
                Key Points
              </h2>
              <ul className="space-y-2">
                {article.keypoints.map((point, index) => (
                  <li
                    key={index}
                    className="flex items-start gap-3 text-sm sm:text-base text-gray-700"
                  >
                    <span className="flex-shrink-0 w-6 h-6 rounded-full bg-blue-100 text-blue-700 flex items-center justify-center text-xs font-semibold">
                      {index + 1}
                    </span>
                    <span>{point}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Impact & Political Score */}
          {(article.impact_score || hasPoliticalScore) && (
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              {/* Impact Score */}
              {article.impact_score && (
                <div className="bg-gray-50 rounded-lg p-4">
                  <div className="flex items-center gap-2 mb-2">
                    <Zap className="w-4 h-4 text-yellow-500" />
                    <span className="text-sm font-semibold text-gray-700">
                      Impact Level
                    </span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span
                      className={`px-3 py-1 rounded-full text-sm font-medium ${
                        article.impact_score === "high"
                          ? "bg-red-100 text-red-700"
                          : article.impact_score === "medium"
                            ? "bg-yellow-100 text-yellow-700"
                            : "bg-green-100 text-green-700"
                      }`}
                    >
                      {article.impact_score === "high"
                        ? "High Impact"
                        : article.impact_score === "medium"
                          ? "Notable"
                          : "Routine"}
                    </span>
                  </div>
                </div>
              )}

              {/* Political Score */}
              {article.political_score !== null &&
                article.political_score !== undefined && (
                  <div className="bg-gray-50 rounded-lg p-4">
                    <div className="text-sm font-semibold text-gray-700 mb-2">
                      Political Leaning
                    </div>
                    <div className="relative h-3 bg-gradient-to-r from-blue-500 via-gray-300 to-red-500 rounded-full">
                      <div
                        className="absolute w-4 h-4 bg-white border-2 border-gray-700 rounded-full -top-0.5 transform -translate-x-1/2"
                        style={{
                          left: `${((article.political_score + 100) / 200) * 100}%`,
                        }}
                      />
                    </div>
                    <div className="flex justify-between text-xs text-gray-500 mt-1">
                      <span>Left</span>
                      <span>Center</span>
                      <span>Right</span>
                    </div>
                  </div>
                )}
            </div>
          )}

          {/* Share Buttons */}
          <div className="pt-4 sm:pt-6 border-t border-gray-200">
            <ShareButtons
              title={article.title}
              url={typeof window !== "undefined" ? window.location.href : ""}
              summary={article.summary}
            />
          </div>

          {/* Source Link */}
          <div className="pt-4 sm:pt-6 border-t border-gray-200">
            <Button asChild className="text-sm sm:text-base">
              <a
                href={article.source_url}
                target="_blank"
                rel="noopener noreferrer"
              >
                <ExternalLink className="w-4 h-4 sm:w-5 sm:h-5" />
                View Full Document on Federal Register
              </a>
            </Button>
          </div>

          {/* Metadata */}
          <div className="pt-4 sm:pt-6 border-t border-gray-200 text-xs sm:text-sm text-gray-500">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4">
              <div>
                <span className="font-semibold">Created:</span>{" "}
                {new Date(article.created_at).toLocaleDateString()}
              </div>
              <div>
                <span className="font-semibold">Last Updated:</span>{" "}
                {new Date(article.updated_at).toLocaleDateString()}
              </div>
            </div>
          </div>
        </div>
      </article>
    </div>
  );
}
