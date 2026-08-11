import { Data } from "effect";

export class TransportError extends Data.TaggedError("TransportError")<{
  readonly cause: unknown;
}> {}

export class HttpError extends Data.TaggedError("HttpError")<{
  readonly status: number;
  readonly code: string;
  readonly message: string;
  readonly requestId?: string;
}> {}

export class DecodeError extends Data.TaggedError("DecodeError")<{
  readonly message: string;
  readonly cause: unknown;
}> {}

export class UnauthorizedError extends Data.TaggedError("UnauthorizedError")<{
  readonly message: string;
}> {}

export class TokenStorageError extends Data.TaggedError("TokenStorageError")<{
  readonly operation: "read" | "write" | "clear";
  readonly cause: unknown;
}> {}

export type ApiFailure = TransportError | HttpError | DecodeError;

export const failureMessage = (failure: unknown): string => {
  if (failure instanceof TokenStorageError) {
    return `Secure token storage ${failure.operation} failed`;
  }
  if (failure instanceof TransportError) return "API request failed";
  if (failure instanceof Error) return failure.message;
  if (typeof failure === "object" && failure !== null && "message" in failure) {
    return String(failure.message);
  }
  return String(failure);
};
