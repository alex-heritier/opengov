import { useState } from "react";
import { Link } from "@tanstack/react-router";
import { FeedList } from "../components/feed/FeedList";
import { Input } from "../components/ui/input";
import { Button } from "../components/ui/button";
import { Search, Bookmark, ThumbsUp, ArrowRight } from "lucide-react";
import { useAuth } from "../hook";
import { WindowChrome } from "../components/ui/window-chrome";

export default function FeedPage() {
  const [searchQuery, setSearchQuery] = useState("");
  const { isAuthenticated } = useAuth();

  return (
    <div className="min-h-screen bg-background text-foreground transition-colors duration-300">
      <div className="w-full max-w-4xl mx-auto px-4 sm:px-6 py-8 sm:py-12">
        {/* Header Section */}
        <header className="mb-10 text-center sm:text-left border-b border-border pb-8">
          <h1 className="text-4xl sm:text-5xl font-chicago tracking-tight mb-3 text-primary flex items-center justify-center sm:justify-start gap-3">
            <span className="text-4xl sm:text-5xl">🇺🇸</span>
            Federal Feed
          </h1>
          <p className="text-xl font-serif italic text-muted-foreground max-w-2xl leading-relaxed">
            Live updates on government actions. Unfiltered. Real-time.
          </p>
        </header>

        <div className="space-y-10">
          {/* Search Bar - Finder Style */}
          <div className="relative group max-w-2xl mx-auto sm:mx-0">
            <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
              <Search className="h-4 w-4 text-muted-foreground/50 group-focus-within:text-primary transition-colors" />
            </div>
            <Input
              type="text"
              placeholder="Search Federal Register entries..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-11 h-12 w-full bg-white border-border/60 rounded-md shadow-sm 
                         font-sans text-base placeholder:font-normal placeholder:text-muted-foreground/40
                         focus:border-primary/50 focus:ring-1 focus:ring-primary/20 
                         transition-all duration-200"
            />
            {/* Keyboard shortcut hint - fun detail */}
            <kbd className="absolute right-3 top-1/2 -translate-y-1/2 hidden sm:block 
                           px-2 py-0.5 text-[10px] font-mono bg-muted rounded border border-border text-muted-foreground">
              ⌘K
            </kbd>
          </div>

          {!isAuthenticated && (
            <WindowChrome title="Access Control" className="bg-card border border-border p-6 sm:p-8 rounded-lg shadow-sm relative overflow-hidden">
              <div className="absolute top-0 right-0 p-4 opacity-5 pointer-events-none">
                {/* Subtle flag motif placeholder if needed, or just abstract shape */}
                <div className="w-32 h-32 bg-primary rounded-full blur-3xl" />
              </div>
              
              <div className="relative z-10 flex flex-col sm:flex-row gap-8 items-start sm:items-center justify-between">
                <div className="space-y-3 max-w-lg">
                  <h3 className="text-xl sm:text-2xl font-chicago tracking-tight text-foreground">
                    Unlock Full Access
                  </h3>
                  <p className="text-muted-foreground leading-relaxed font-sans">
                    Sign in to bookmark regulations, track specific agencies, and curate your personalized feed.
                  </p>
                  <div className="flex gap-6 text-sm font-medium text-muted-foreground/80 pt-2 font-chicago">
                    <div className="flex items-center gap-2">
                      <Bookmark className="w-4 h-4 text-primary" />
                      <span>Save Articles</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <ThumbsUp className="w-4 h-4 text-primary" />
                      <span>Track Issues</span>
                    </div>
                  </div>
                </div>
                <div className="flex flex-col sm:flex-row gap-3 w-full sm:w-auto min-w-[140px]">
                  <Button 
                    asChild 
                    size="lg" 
                    className="rounded-md font-chicago h-11 shadow-sm hover-lift"
                  >
                    <Link to="/login">Sign In <ArrowRight className="ml-2 w-4 h-4" /></Link>
                  </Button>
                  <Button 
                    asChild 
                    variant="outline" 
                    size="lg" 
                    className="rounded-md border-border text-foreground hover:bg-accent hover:text-accent-foreground font-chicago h-11 hover-lift"
                  >
                    <Link to="/feed">Guest Access</Link>
                  </Button>
                </div>
              </div>
            </WindowChrome>
          )}

          {/* Feed List */}
          <main>
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-2xl font-chicago text-primary border-b-2 border-primary/20 pb-1">
                Latest Entries
              </h2>
            </div>
            <FeedList />
          </main>
        </div>
      </div>
    </div>
  );
}
