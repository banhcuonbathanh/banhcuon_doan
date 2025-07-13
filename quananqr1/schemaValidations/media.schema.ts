import { z } from "zod";

export const UploadImageRes = z.object({
  message: z.string(),
  filename: z.string(),
  path: z.string()
});

export type UploadImageResType = z.TypeOf<typeof UploadImageRes>;
