import type { components } from "./generated/api-types";
import type {
  APIError,
  AdminPingResponse,
  LandingInfo,
  MagicLinkInput,
  MessageResponse,
  MobileGoogleInput,
  MobileVerifyInput,
  RefreshInput,
  TokenPair,
  UpdateProfileInput,
  User,
} from "./contracts";

type Assert<T extends true> = T;
type Compatible<A, B> = A extends B ? (B extends A ? true : false) : false;
type Generated = components["schemas"];

type ContractAssertions = [
  Assert<Compatible<APIError, Generated["APIError"]>>,
  Assert<Compatible<AdminPingResponse, Generated["AdminPingResponse"]>>,
  Assert<Compatible<LandingInfo, Generated["LandingInfo"]>>,
  Assert<Compatible<MagicLinkInput, Generated["MagicLinkInputBody"]>>,
  Assert<Compatible<MessageResponse, Generated["MessageResponse"]>>,
  Assert<Compatible<MobileGoogleInput, Generated["MobileGoogleInputBody"]>>,
  Assert<Compatible<MobileVerifyInput, Generated["MobileVerifyInputBody"]>>,
  Assert<Compatible<RefreshInput, Generated["RefreshInputBody"]>>,
  Assert<Compatible<TokenPair, Generated["TokenPair"]>>,
  Assert<Compatible<UpdateProfileInput, Generated["UpdateProfileInputBody"]>>,
  Assert<Compatible<User, Generated["User"]>>,
];

export type ContractsMatchGeneratedOpenAPI = ContractAssertions;
