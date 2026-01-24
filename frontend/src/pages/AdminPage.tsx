import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import apiClient from "@/api/client";

interface ScraperRun {
  id: number;
  started_at: string;
  completed_at: string | null;
  processed_count: number;
  skipped_count: number;
  error_count: number;
  success: boolean;
  error_message: string | null;
  duration_seconds: number | null;
}

interface ScraperRunListResponse {
  runs: ScraperRun[];
  total: number;
}

export default function AdminPage() {
  const [limit, setLimit] = useState(10);

  // Fetch scraper runs
  const { data, isLoading, error } = useQuery<ScraperRunListResponse>({
    queryKey: ["admin", "scraper-runs", limit],
    queryFn: async () => {
      const response = await apiClient.get("/api/admin/scraper-runs", {
        params: { limit },
      });
      return response.data;
    },
  });

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString();
  };

  const formatDuration = (seconds: number | null) => {
    if (seconds === null) return "In progress";
    if (seconds < 60) return `${Math.round(seconds)}s`;
    return `${(seconds / 60).toFixed(1)}m`;
  };

  const getStatusBadge = (run: ScraperRun) => {
    if (!run.completed_at) {
      return (
        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
          Running
        </span>
      );
    }
    return run.success ? (
      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
        Success
      </span>
    ) : (
      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
        Failed
      </span>
    );
  };

  return (
    <div className="space-y-6 p-6">
      <div>
        <h1 className="text-3xl font-bold">Admin Dashboard</h1>
        <p className="text-gray-500 mt-2">
          Manage scraper jobs and monitor system health
        </p>
      </div>

      {/* Scraper Runs Table */}
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold mb-2">Recent Scraper Runs</h2>
        <p className="text-gray-600 text-sm mb-4">
          Last {limit} jobs {data && `(${data.total} total)`}
        </p>

        {isLoading ? (
          <div className="flex justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
          </div>
        ) : error ? (
          <div className="text-red-500 text-center py-8">
            Failed to load scraper runs
          </div>
        ) : !data?.runs.length ? (
          <div className="text-gray-500 text-center py-8">
            No scraper runs found
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full border-collapse">
              <thead>
                <tr className="border-b border-gray-200">
                  <th className="text-left px-4 py-3 text-sm font-semibold text-gray-700">
                    ID
                  </th>
                  <th className="text-left px-4 py-3 text-sm font-semibold text-gray-700">
                    Started
                  </th>
                  <th className="text-left px-4 py-3 text-sm font-semibold text-gray-700">
                    Duration
                  </th>
                  <th className="text-left px-4 py-3 text-sm font-semibold text-gray-700">
                    Processed
                  </th>
                  <th className="text-left px-4 py-3 text-sm font-semibold text-gray-700">
                    Skipped
                  </th>
                  <th className="text-left px-4 py-3 text-sm font-semibold text-gray-700">
                    Errors
                  </th>
                  <th className="text-left px-4 py-3 text-sm font-semibold text-gray-700">
                    Status
                  </th>
                  <th className="text-left px-4 py-3 text-sm font-semibold text-gray-700">
                    Error Message
                  </th>
                </tr>
              </thead>
              <tbody>
                {data.runs.map((run) => (
                  <tr
                    key={run.id}
                    className="border-b border-gray-200 hover:bg-gray-50"
                  >
                    <td className="px-4 py-3 text-sm font-medium">{run.id}</td>
                    <td className="px-4 py-3 text-sm text-gray-600">
                      {formatDate(run.started_at)}
                    </td>
                    <td className="px-4 py-3 text-sm">
                      {formatDuration(run.duration_seconds)}
                    </td>
                    <td className="px-4 py-3 text-sm">
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                        {run.processed_count}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm">
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                        {run.skipped_count}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-sm">
                      {run.error_count > 0 ? (
                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
                          {run.error_count}
                        </span>
                      ) : (
                        <span className="text-gray-400">0</span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-sm">{getStatusBadge(run)}</td>
                    <td className="px-4 py-3 text-sm text-gray-600 max-w-xs truncate">
                      {run.error_message || "-"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {data && data.total > limit && (
          <div className="mt-4 text-center">
            <button
              onClick={() => setLimit(Math.min(limit + 10, 50))}
              disabled={limit >= 50}
              className="inline-flex items-center px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Load More
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
