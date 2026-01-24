export type Environment = "development" | "production" | "test" | "preview";

export function useEnvironment(): {
  env: Environment;
  isDevelopment: boolean;
  isProduction: boolean;
  isTest: boolean;
  isPreview: boolean;
} {
  const mode = import.meta.env.MODE as Environment;

  return {
    env: mode,
    isDevelopment: mode === "development",
    isProduction: mode === "production",
    isTest: mode === "test",
    isPreview: mode === "preview",
  };
}
