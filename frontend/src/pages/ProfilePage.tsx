import React, { useState, useEffect } from "react";
import { useProfile, useAuth } from "@/hook";
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
  LogOut,
  Flag,
  Vote,
  MapPin,
} from "lucide-react";

const POLITICAL_LEANINGS = [
  { value: "democrat", label: "Democrat" },
  { value: "republican", label: "Republican" },
  { value: "libertarian", label: "Libertarian" },
  { value: "maga", label: "MAGA" },
  { value: "america_first", label: "America First" },
  { value: "socialist", label: "Socialist" },
];

const US_STATES = [
  { value: "AL", label: "Alabama" },
  { value: "AK", label: "Alaska" },
  { value: "AZ", label: "Arizona" },
  { value: "AR", label: "Arkansas" },
  { value: "CA", label: "California" },
  { value: "CO", label: "Colorado" },
  { value: "CT", label: "Connecticut" },
  { value: "DE", label: "Delaware" },
  { value: "FL", label: "Florida" },
  { value: "GA", label: "Georgia" },
  { value: "HI", label: "Hawaii" },
  { value: "ID", label: "Idaho" },
  { value: "IL", label: "Illinois" },
  { value: "IN", label: "Indiana" },
  { value: "IA", label: "Iowa" },
  { value: "KS", label: "Kansas" },
  { value: "KY", label: "Kentucky" },
  { value: "LA", label: "Louisiana" },
  { value: "ME", label: "Maine" },
  { value: "MD", label: "Maryland" },
  { value: "MA", label: "Massachusetts" },
  { value: "MI", label: "Michigan" },
  { value: "MN", label: "Minnesota" },
  { value: "MS", label: "Mississippi" },
  { value: "MO", label: "Missouri" },
  { value: "MT", label: "Montana" },
  { value: "NE", label: "Nebraska" },
  { value: "NV", label: "Nevada" },
  { value: "NH", label: "New Hampshire" },
  { value: "NJ", label: "New Jersey" },
  { value: "NM", label: "New Mexico" },
  { value: "NY", label: "New York" },
  { value: "NC", label: "North Carolina" },
  { value: "ND", label: "North Dakota" },
  { value: "OH", label: "Ohio" },
  { value: "OK", label: "Oklahoma" },
  { value: "OR", label: "Oregon" },
  { value: "PA", label: "Pennsylvania" },
  { value: "RI", label: "Rhode Island" },
  { value: "SC", label: "South Carolina" },
  { value: "SD", label: "South Dakota" },
  { value: "TN", label: "Tennessee" },
  { value: "TX", label: "Texas" },
  { value: "UT", label: "Utah" },
  { value: "VT", label: "Vermont" },
  { value: "VA", label: "Virginia" },
  { value: "WA", label: "Washington" },
  { value: "WV", label: "West Virginia" },
  { value: "WI", label: "Wisconsin" },
  { value: "WY", label: "Wyoming" },
];

export default function ProfilePage() {
  const { user, updateUser, logout } = useAuth();
  const { updateProfileAsync, isUpdating } = useProfile();
  const [politicalLeaning, setPoliticalLeaning] = useState("");
  const [state, setState] = useState("");
  const [showSuccess, setShowSuccess] = useState(false);

  useEffect(() => {
    if (user) {
      setPoliticalLeaning(user.political_leaning || "");
      setState(user.state || "");
    }
  }, [user]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      const updatedUser = await updateProfileAsync({
        political_leaning:
          politicalLeaning === "" || politicalLeaning === "prefer-not-to-say"
            ? null
            : politicalLeaning,
        state: state === "" || state === "prefer-not-to-say" ? null : state,
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
      <div className="min-h-screen bg-background p-6 flex items-center justify-center">
        <div className="max-w-md w-full bg-card border border-border p-8 rounded-none shadow-sm text-center">
          <Vote className="w-12 h-12 text-primary mx-auto mb-4" />
          <h2 className="text-2xl font-chicago mb-2">Authentication Required</h2>
          <p className="text-muted-foreground mb-6">Please log in to view your citizen profile.</p>
          <Button asChild>
            <a href="/login" className="font-chicago">Sign In</a>
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background text-foreground transition-colors duration-300 pb-20">
      <div className="w-full max-w-4xl mx-auto px-4 sm:px-6 py-8 sm:py-12 space-y-8">
        
        {/* Header */}
        <header className="border-b-2 border-primary/20 pb-6 mb-8 flex items-center gap-4">
          <div className="bg-primary/10 p-3 rounded-full">
            <UserIcon className="w-8 h-8 text-primary" />
          </div>
          <div>
            <h1 className="text-3xl sm:text-4xl font-chicago tracking-tight text-primary">
              Citizen Profile
            </h1>
            <p className="text-lg font-serif italic text-muted-foreground">
              Manage your registration and preferences.
            </p>
          </div>
        </header>

        <div className="grid gap-8 md:grid-cols-2">
          
          {/* Left Column: ID Card & Actions */}
          <div className="space-y-6">
            {/* ID Card Style Info */}
            <div className="bg-white border border-border rounded-md shadow-sm overflow-hidden">
              {/* Cleaner header - no gradient, just stripes */}
              <div className="bg-primary/5 border-b border-border px-4 py-3 flex justify-between items-center">
                <span className="font-mono text-[10px] uppercase tracking-widest text-primary/70 font-bold">
                  Citizen ID • {user.id}
                </span>
                <Flag className="w-4 h-4 text-primary/30" />
              </div>
              
              <div className="p-6">
                <div className="flex items-start gap-4">
                  {/* Avatar with subtle ring */}
                  <div className="w-20 h-20 rounded-full bg-secondary border-2 border-border overflow-hidden flex-shrink-0">
                    {user.picture_url ? (
                      <img src={user.picture_url} alt="" className="w-full h-full object-cover" />
                    ) : (
                      <div className="w-full h-full flex items-center justify-center bg-primary/5">
                        <UserIcon className="w-8 h-8 text-primary/40" />
                      </div>
                    )}
                  </div>
                  
                  <div className="flex-1 min-w-0">
                    <h2 className="font-serif text-2xl italic text-foreground mb-1 truncate">
                      {user.name || "Anonymous Citizen"}
                    </h2>
                    <p className="text-sm text-muted-foreground font-mono truncate">{user.email}</p>
                    
                    {/* Meta grid - cleaner */}
                    <div className="grid grid-cols-2 gap-4 mt-4 pt-4 border-t border-border/50">
                      <div>
                        <span className="block text-[10px] uppercase tracking-wider text-muted-foreground font-semibold mb-0.5">Member Since</span>
                        <span className="text-sm font-medium tabular-nums">
                          {new Date(user.created_at).toLocaleDateString("en-US", { month: "short", year: "numeric" })}
                        </span>
                      </div>
                      <div>
                        <span className="block text-[10px] uppercase tracking-wider text-muted-foreground font-semibold mb-0.5">Status</span>
                        <div className="flex items-center gap-1.5">
                          <span className="w-1.5 h-1.5 rounded-full bg-success" />
                          <span className="text-sm font-medium">Active</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {/* Account Actions */}
            <Card className="border-border shadow-sm">
              <CardHeader className="pb-3">
                <CardTitle className="text-lg font-chicago">Session Management</CardTitle>
              </CardHeader>
              <CardContent>
                <Button
                  onClick={logout}
                  variant="destructive"
                  className="w-full flex items-center justify-center gap-2 font-chicago"
                >
                  <LogOut className="w-4 h-4" />
                  Sign Out
                </Button>
              </CardContent>
            </Card>
          </div>

          {/* Right Column: Preferences Form */}
          <div className="space-y-6">
            <Card className="border-2 border-primary/20 shadow-md bg-card relative overflow-hidden">
              <div className="absolute top-0 right-0 p-4 opacity-5 pointer-events-none">
                <Vote className="w-24 h-24 rotate-12" />
              </div>

              <CardHeader className="bg-primary/5 border-b border-border pb-6">
                <div className="flex items-center gap-2 mb-1 text-primary">
                  <Vote className="w-5 h-5" />
                  <span className="text-xs font-bold uppercase tracking-widest font-chicago">Voter Preferences</span>
                </div>
                <CardTitle className="text-2xl font-serif text-foreground">Civic Profile</CardTitle>
                <CardDescription className="text-muted-foreground font-serif italic">
                  Personalize your feed based on your location and political alignment.
                </CardDescription>
              </CardHeader>
              
              <CardContent className="pt-6">
                <form onSubmit={handleSubmit} className="space-y-6">
                  <div className="space-y-3">
                    <Label htmlFor="state" className="text-sm font-bold uppercase tracking-wider text-muted-foreground flex items-center gap-2">
                      <MapPin className="w-4 h-4" />
                      State / Territory
                    </Label>
                    <Select value={state} onValueChange={setState}>
                      <SelectTrigger id="state" className="h-12 bg-background border-input text-lg font-medium focus:ring-primary">
                        <SelectValue placeholder="Select your state" />
                      </SelectTrigger>
                      <SelectContent className="max-h-[300px]">
                        <SelectItem value="prefer-not-to-say" className="text-muted-foreground italic">
                          Prefer not to say
                        </SelectItem>
                        {US_STATES.map((option) => (
                          <SelectItem key={option.value} value={option.value} className="font-medium">
                            {option.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-3">
                    <Label htmlFor="political-leaning" className="text-sm font-bold uppercase tracking-wider text-muted-foreground flex items-center gap-2">
                      <Flag className="w-4 h-4" />
                      Political Alignment
                    </Label>
                    <Select
                      value={politicalLeaning}
                      onValueChange={setPoliticalLeaning}
                    >
                      <SelectTrigger id="political-leaning" className="h-12 bg-background border-input text-lg font-medium focus:ring-primary">
                        <SelectValue placeholder="Select your political leaning" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="prefer-not-to-say" className="text-muted-foreground italic">
                          Prefer not to say
                        </SelectItem>
                        {POLITICAL_LEANINGS.map((option) => (
                          <SelectItem key={option.value} value={option.value} className="font-medium">
                            {option.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <p className="text-xs text-muted-foreground ml-1">
                      * Used to highlight relevant regulations and generate personalized summaries.
                    </p>
                  </div>

                  {showSuccess && (
                    <Alert className="bg-green-50 border-green-200 animate-in fade-in slide-in-from-top-2">
                      <CheckCircle2 className="h-4 w-4 text-green-600" />
                      <AlertDescription className="text-green-800 font-medium">
                        Preferences updated successfully.
                      </AlertDescription>
                    </Alert>
                  )}

                  <div className="pt-4">
                    <Button
                      type="submit"
                      disabled={isUpdating}
                      className="w-full h-11 font-chicago text-base shadow-sm"
                    >
                      {isUpdating ? "Updating Records..." : "Update Records"}
                    </Button>
                  </div>
                </form>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}
