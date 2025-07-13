import { create } from "zustand";
import { persist } from "zustand/middleware";

import { UploadImageResType } from "@/schemaValidations/media.schema";
import { useApiStore } from "../api/api-controller";
import envConfig from "@/config";
interface MediaStore {
  isUploading: boolean;
  error: string | null;
  uploadMedia: (file: File, uploadPath: string) => Promise<UploadImageResType>;

}

export const useMediaStore = create<MediaStore>()(
  persist(
    (set, get) => ({
      isUploading: false,
      error: null,
      uploadMedia: async (file: File, uploadPath: string) => {
        const link =
          envConfig.NEXT_PUBLIC_API_ENDPOINT +
          envConfig.NEXT_PUBLIC_Image_Upload;

        const { http } = useApiStore.getState();
        set({ isUploading: true, error: null });

        try {
          const formData = new FormData();
          formData.append("image", file); // Use "image" as the key to match the backend
          formData.append("path", uploadPath); // Add the upload path to the form data

          const response = await http.post<UploadImageResType>(link, formData);
          set({ isUploading: false });

          console.log(
            "quananqr1/zusstand/media/usemediastore.ts response.data.data",
            response.data
          );
          return response.data; // Adjust as per your actual response structure
        } catch (error) {
          set({ isUploading: false, error: "Failed to upload media" });
          throw error;
        }
      }
    }),
    {
      name: "media-storage",
      skipHydration: true
    }
  )
);

// Custom hook for media upload mutation
export const useUploadMediaMutation = () => {
  const { uploadMedia, isUploading, error } = useMediaStore();
  return {
    mutateAsync: uploadMedia,
    isPending: isUploading,
    error
  };
};
