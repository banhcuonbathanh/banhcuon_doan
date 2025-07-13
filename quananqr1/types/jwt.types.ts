import { Role, TokenType } from "@/constants/type";

export type TokenTypeValue = (typeof TokenType)[keyof typeof TokenType];
export type RoleType = (typeof Role)[keyof typeof Role];
export interface TokenPayload {
  id: number;
  role: RoleType;
  tokenType: TokenTypeValue;
  exp: number;
  iat: number;
  email: string;
}

export interface TableTokenPayload {
  iat: number;
  number: number;
  tokenType: (typeof TokenType)["TableToken"];
}
