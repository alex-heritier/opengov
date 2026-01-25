export const SourceType = {
  FEDERAL_REGISTER: "federal_register",
} as const;

export type SourceType = (typeof SourceType)[keyof typeof SourceType];
