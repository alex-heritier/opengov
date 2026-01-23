import { useMutation, useQueryClient } from "@tanstack/react-query";
import client from "@/api/client";

export interface UserUpdate {
  name?: string | null;
  picture_url?: string | null;
  political_leaning?: string | null;
  state?: string | null;
}

export function useUpdateProfileMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (updates: UserUpdate) => {
      const { data } = await client.patch("/api/users/me", updates);
      return data;
    },
    onSuccess: (data) => {
      // Update the user in auth store by invalidating the current user query
      queryClient.invalidateQueries({ queryKey: ["currentUser"] });
      // Also trigger a refetch of the user from the auth endpoint
      return data;
    },
  });
}
