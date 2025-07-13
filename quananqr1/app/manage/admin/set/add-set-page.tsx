"use client";

import React, { useState, useRef, useMemo, useEffect } from "react";
import { useFieldArray, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Checkbox } from "@/components/ui/checkbox";

import { Button } from "@/components/ui/button";
import {
  Form,
  FormField,
  FormItem,
  FormControl,
  FormMessage,
  FormLabel,
  FormDescription
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from "@/components/ui/select";
import { Avatar, AvatarImage, AvatarFallback } from "@/components/ui/avatar";
import { Upload } from "lucide-react";
import envConfig from "@/config";
import { DishStatus, DishStatusValues } from "@/constants/type";
import { handleErrorApi, getVietnameseDishStatus } from "@/lib/utils";
import {
  CreateDishBodyType,
  CreateDishBody
} from "@/schemaValidations/dish.schema";
import { useDishStore } from "@/zusstand/dished/dished-controller";
import { useMediaStore } from "@/zusstand/media/usemediastore";
import { SetCreateBodyInterface } from "@/schemaValidations/interface/types_set";
import { z } from "zod";
const SetCreateBody = z.object({
  image: z.string(),
  name: z.string().min(1, "Name is required"),
  description: z.string().optional(),
  dishes: z.array(
    z.object({
      dishId: z.number(),
      quantity: z.number().min(1),
      dish: z.any() // You might want to define a more specific schema for DishInterface
    })
  ),
  userId: z.number(),
  created_at: z.string(),
  updated_at: z.string(),
  is_favourite: z.boolean(),
  like_by: z.array(z.number()).nullable(),
  is_public: z.boolean()
});
export default function AddSetPage() {
  const [file, setFile] = useState<File | null>(null);
  const { uploadMedia, isUploading: isUploadingMedia } = useMediaStore();
  const { dishes, addSet, isLoading: isAddingSet } = useDishStore();

  const imageInputRef = useRef<HTMLInputElement | null>(null);
  const form = useForm<SetCreateBodyInterface>({
    resolver: zodResolver(SetCreateBody),
    defaultValues: {
      image: "",
      name: "",
      description: "",
      dishes: [],
      userId: 1, // You might want to get this from the current user's session
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      is_favourite: false,
      like_by: null,
      is_public: false
    }
  });

  const image = form.watch("image");
  const name = form.watch("name");
  const previewAvatarFromFile = useMemo(() => {
    if (file) {
      return URL.createObjectURL(file);
    }
    return image;
  }, [file, image]);

  const reset = () => {
    form.reset();
    setFile(null);
  };

  const onSubmit = async (values: SetCreateBodyInterface) => {
    console.log(
      "quananqr1/app/admin/set/add-set-page.tsx onSubmit SetCreateBodyInterface ",
      values
    );

    if (isAddingSet || isUploadingMedia) return;
    try {
      let body = values;
      if (file) {
        const imageUrl = await uploadMedia(
          file,
          envConfig.NEXT_PUBLIC_Folder1_BE + values.name
        );

        body = {
          ...values,
          image:
            envConfig.NEXT_PUBLIC_API_ENDPOINT +
            envConfig.NEXT_PUBLIC_Upload +
            imageUrl.path
        };
      }

      const result = await addSet(body);

      console.log("quananqr1/app/admin/set/add-set-page.tsx onSubmit done");
      reset();
      // You might want to add some success feedback here
    } catch (error) {
      handleErrorApi({
        error,
        setError: form.setError
      });
    }
  };
  const { fields, append, remove, update } = useFieldArray({
    control: form.control,
    name: "dishes"
  });
  return (
    <div className="container mx-auto">
      <h1 className="text-2xl font-bold mb-4">Thêm Set Món Ăn</h1>
      <Form {...form}>
        <form
          noValidate
          className="grid auto-rows-max items-start gap-4 md:gap-8"
          onSubmit={form.handleSubmit(onSubmit, (e) => {
            console.log(e);
          })}
          onReset={reset}
        >
          <div className="grid gap-4 py-4">
            <FormField
              control={form.control}
              name="image"
              render={({ field }) => (
                <FormItem>
                  <div className="flex gap-2 items-start justify-start">
                    <Avatar className="aspect-square w-[100px] h-[100px] rounded-md object-cover">
                      <AvatarImage src={previewAvatarFromFile} />
                      <AvatarFallback className="rounded-none">
                        {name || "Ảnh set món ăn"}
                      </AvatarFallback>
                    </Avatar>
                    <input
                      type="file"
                      accept="image/*"
                      ref={imageInputRef}
                      onChange={(e) => {
                        const file = e.target.files?.[0];
                        if (file) {
                          setFile(file);
                          field.onChange("http://localhost:3000/" + file.name);
                        }
                      }}
                      className="hidden"
                    />
                    <button
                      className="flex aspect-square w-[100px] items-center justify-center rounded-md border border-dashed"
                      type="button"
                      onClick={() => imageInputRef.current?.click()}
                    >
                      <Upload className="h-4 w-4 text-muted-foreground" />
                      <span className="sr-only">Upload</span>
                    </button>
                  </div>
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <div className="grid grid-cols-4 items-center justify-items-start gap-4">
                    <Label htmlFor="name">Tên set món ăn</Label>
                    <div className="col-span-3 w-full space-y-2">
                      <Input id="name" className="w-full" {...field} />
                      <FormMessage />
                    </div>
                  </div>
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <div className="grid grid-cols-4 items-center justify-items-start gap-4">
                    <Label htmlFor="description">Mô tả set món ăn</Label>
                    <div className="col-span-3 w-full space-y-2">
                      <Textarea
                        id="description"
                        className="w-full"
                        {...field}
                      />
                      <FormMessage />
                    </div>
                  </div>
                </FormItem>
              )}
            />

            {/* Add a component for managing dishes in the set */}
            {/* This could be a custom component that allows adding/removing dishes and setting quantities */}
            {/* For example: <DishSelector control={form.control} name="dishes" availableDishes={dishes} /> */}

            <FormField
              control={form.control}
              name="is_public"
              render={({ field }) => (
                <FormItem className="flex flex-row items-start space-x-3 space-y-0 rounded-md border p-4">
                  <FormControl>
                    <Checkbox
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  </FormControl>
                  <div className="space-y-1 leading-none">
                    <FormLabel>Public</FormLabel>
                    <FormDescription>
                      Make this set visible to other users
                    </FormDescription>
                  </div>
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="is_favourite"
              render={({ field }) => (
                <FormItem className="flex flex-row items-start space-x-3 space-y-0 rounded-md border p-4">
                  <FormControl>
                    <Checkbox
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  </FormControl>
                  <div className="space-y-1 leading-none">
                    <FormLabel>Yêu thích</FormLabel>
                    <FormDescription>
                      Mark this set as a favorite
                    </FormDescription>
                  </div>
                </FormItem>
              )}
            />
          </div>
          {/* add product */}

          <div className="grid gap-4 py-4">
            <h2 className="text-xl font-semibold">Chọn món ăn cho set</h2>
            {dishes.map((dish) => {
              const fieldIndex = fields.findIndex(
                (field) => field.dishId === dish.id
              );
              const isSelected = fieldIndex !== -1;

              return (
                <div
                  key={dish.id}
                  className="flex items-center space-x-4 border p-4 rounded-md"
                >
                  <Avatar className="h-16 w-16">
                    <AvatarImage src={dish.image} alt={dish.name} />
                    <AvatarFallback>
                      {dish.name.slice(0, 2).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex-grow">
                    <h3 className="font-semibold">{dish.name}</h3>
                    {dish.set_id && (
                      <p className="text-sm text-gray-500">
                        Món ăn này đã thuộc một set khác
                      </p>
                    )}
                  </div>
                  <div className="flex items-center space-x-2">
                    <Checkbox
                      checked={isSelected}
                      onCheckedChange={(checked) => {
                        if (checked) {
                          append({ dishId: dish.id, quantity: 1, dish: dish });
                        } else {
                          remove(fieldIndex);
                        }
                      }}
                      disabled={!!dish.set_id}
                    />
                    {isSelected && (
                      <Input
                        type="number"
                        min="1"
                        className="w-16"
                        value={fields[fieldIndex].quantity}
                        onChange={(e) => {
                          const newQuantity = parseInt(e.target.value, 10);
                          if (!isNaN(newQuantity) && newQuantity > 0) {
                            update(fieldIndex, {
                              ...fields[fieldIndex],
                              quantity: newQuantity
                            });
                          }
                        }}
                      />
                    )}
                  </div>
                </div>
              );
            })}
          </div>
          {/* add product */}
          <div className="flex justify-end space-x-4">
            <Button type="reset" variant="outline">
              Hủy
            </Button>
            <Button type="submit">Thêm Set</Button>
          </div>
        </form>
      </Form>
    </div>
  );
}
