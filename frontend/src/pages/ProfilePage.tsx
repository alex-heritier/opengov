import React, { useState } from "react";
import { useAuthStore } from "@/store/authStore";
import { useAuth, useProfile } from "@/hook";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  CheckCircle2,
  User as UserIcon,
  Mail,
  Calendar,
  LogOut,
} from "lucide-react";

const POLITICAL_LEANINGS = [
  { value: "democrat", label: "Democrat" },
  { value: "republican", label: "Republican" },
  { value: "libertarian", label: "Libertarian" },
  { value: "maga", label: "MAGA" },
  { value: "america_first", label: "America First" },
  { value: "socialist", label: "Socialist" },
];

export default function ProfilePage() {
  const { user, updateUser } = useAuthStore();
  const { logout } = useAuth();
  const { updateProfileAsync, isUpdating } = useProfile();
  const [politicalLeaning, setPoliticalLeaning] = useState<string | undefined>(
    user?.political_leaning || undefined,
  );
  const [showSuccess, setShowSuccess] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      const updatedUser = await updateProfileAsync({
        political_leaning:
          politicalLeaning === "prefer-not-to-say"
            ? null
            : politicalLeaning || null,
      });
      // Update the user in the auth store
      updateUser(updatedUser);
      setShowSuccess(true);
      setTimeout(() => setShowSuccess(false), 3000);
    } catch (error) {
      console.error("Failed to update profile:", error);
    }
  };

  if (!user) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Alert>
          <AlertDescription>
            Please log in to view your profile.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8 max-w-2xl">
      <h1 className="text-3xl font-bold mb-6">Profile</h1>

      {/* User Info Card */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle>Account Information</CardTitle>
          <CardDescription>Your basic account details</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-3">
            <Mail className="w-5 h-5 text-gray-500" />
            <div>
              <p className="text-sm text-gray-500">Email</p>
              <p className="font-medium">{user.email}</p>
            </div>
          </div>

          {user.name && (
            <div className="flex items-center gap-3">
              <UserIcon className="w-5 h-5 text-gray-500" />
              <div>
                <p className="text-sm text-gray-500">Name</p>
                <p className="font-medium">{user.name}</p>
              </div>
            </div>
          )}

          <div className="flex items-center gap-3">
            <Calendar className="w-5 h-5 text-gray-500" />
            <div>
              <p className="text-sm text-gray-500">Member Since</p>
              <p className="font-medium">
                {new Date(user.created_at).toLocaleDateString("en-US", {
                  year: "numeric",
                  month: "long",
                  day: "numeric",
                })}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Political Leaning Form */}
      <Card>
        <CardHeader>
          <CardTitle>Political Preferences</CardTitle>
          <CardDescription>
            Help us personalize your experience by sharing your political
            leaning
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="political-leaning">Political Leaning</Label>
              <Select
                value={politicalLeaning}
                onValueChange={setPoliticalLeaning}
              >
                <SelectTrigger id="political-leaning">
                  <SelectValue placeholder="Select your political leaning" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="prefer-not-to-say">
                    Prefer not to say
                  </SelectItem>
                  {POLITICAL_LEANINGS.map((option) => (
                    <SelectItem key={option.value} value={option.value}>
                      {option.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {showSuccess && (
              <Alert className="bg-green-50 border-green-200">
                <CheckCircle2 className="h-4 w-4 text-green-600" />
                <AlertDescription className="text-green-800">
                  Profile updated successfully!
                </AlertDescription>
              </Alert>
            )}

            <Button
              type="submit"
              disabled={isUpdating}
              className="w-full sm:w-auto"
            >
              {isUpdating ? "Saving..." : "Save Changes"}
            </Button>
          </form>
        </CardContent>
      </Card>

      {/* Account Actions */}
      <Card className="mt-6">
        <CardHeader>
          <CardTitle>Account Actions</CardTitle>
          <CardDescription>Manage your account session</CardDescription>
        </CardHeader>
        <CardContent>
          <Button
            onClick={logout}
            variant="destructive"
            className="w-full sm:w-auto flex items-center gap-2"
          >
            <LogOut className="w-4 h-4" />
            Log Out
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
