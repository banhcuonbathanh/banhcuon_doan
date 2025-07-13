// "use client";

// import React, { useEffect, useRef, useState } from "react";
// import { useForm } from "react-hook-form";
// import { zodResolver } from "@hookform/resolvers/zod";

// import { Info } from "lucide-react";
// import {
//   Dialog,
//   DialogContent,
//   DialogTrigger,
//   DialogTitle,
//   DialogHeader
// } from "@/components/ui/dialog";
// import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
// import { Button } from "@/components/ui/button";
// import { Input } from "@/components/ui/input";
// import { Label } from "@/components/ui/label";

// import { handleErrorApi } from "@/lib/utils";
// import { useAuthStore } from "../../../../auth/controller/auth-controller";
// import {
//   RegisterBodyType,
//   RegisterBody
// } from "../../../../auth/domain/auth.schema";

// const RegisterDialog = () => {
//   const [open, setOpen] = useState(true);

//   const renderCount = useRef(0);
//   renderCount.current += 1;

//   if (renderCount.current === 1) {
//     console.log(
//       "quananqr1/app/(public)/public-component/register-dialog.tsx renderCount.current logic in if is execute "
//     );
//     setOpen(false);
//     setOpen(false);
//     console.log(
//       "quananqr1/app/(public)/public-component/register-dialog.tsx renderCount.current logic in if is execute done "
//     );
//   }

//   const { register, openLoginDialog } = useAuthStore();
//   const form = useForm<RegisterBodyType>({
//     resolver: zodResolver(RegisterBody),
//     defaultValues: {
//       name: "",
//       email: "",
//       password: "",
//       role: "",
//       phone: "",
//       image: "",
//       address: "",
//       created_at: new Date().toISOString(),
//       updated_at: new Date().toISOString()
//     }
//   });

//   const onSubmit = async (data: RegisterBodyType) => {
//     try {
//       console.log(
//         "onSubmit register form quananqr1/app/(public)/public-component/register-dialog.tsx data",
//         data
//       );

//       await register({
//         name: data.name,
//         email: data.email,
//         password: data.password,
//         role: "Guest",
//         phone: data.phone,
//         image: data.image,
//         address: data.address,
//         created_at: data.created_at,
//         updated_at: data.updated_at
//       });
//       setOpen(false);
//       openLoginDialog();
//     } catch (error: any) {
//       console.log("Error during registration: ", error);
//       handleErrorApi({
//         error,
//         setError: form.setError
//       });
//     }
//   };

//   return (
//     <Dialog open={open} onOpenChange={setOpen}>
//       <DialogTrigger asChild>
//         <Button>Đăng ký ewrt</Button>
//       </DialogTrigger>
//       <DialogContent className="sm:max-w-[425px] bg-white dark:bg-gray-800 shadow-lg">
//         <DialogHeader>
//           <DialogTitle className="text-2xl font-semibold">Đăng ký</DialogTitle>
//         </DialogHeader>
//         <p className="text-sm text-gray-500 dark:text-gray-400">
//           Điền thông tin của bạn để tạo tài khoản mới
//         </p>
//         <Form {...form}>
//           <form
//             className="space-y-4 w-full"
//             noValidate
//             onSubmit={form.handleSubmit(onSubmit, (err) => {
//               console.log(
//                 "Registration err onSubmit: quananqr1/app/(public)/public-component/register-dialog.tsx",
//                 err
//               );
//             })}
//           >
//             <div className="grid gap-4">
//               <FormField
//                 control={form.control}
//                 name="name"
//                 render={({ field }) => (
//                   <FormItem>
//                     <div className="grid gap-2">
//                       <Label htmlFor="name">Họ và tên</Label>
//                       <Input
//                         id="name"
//                         type="text"
//                         placeholder="Alice Johnson"
//                         required
//                         className="border-2 border-gray-300 dark:border-gray-600"
//                         {...field}
//                       />
//                       <FormMessage />
//                     </div>
//                   </FormItem>
//                 )}
//               />
//               <FormField
//                 control={form.control}
//                 name="email"
//                 render={({ field }) => (
//                   <FormItem>
//                     <div className="grid gap-2">
//                       <Label htmlFor="email">Email</Label>
//                       <Input
//                         id="email"
//                         type="email"
//                         placeholder="alice.johnson@example.com"
//                         required
//                         className="border-2 border-gray-300 dark:border-gray-600"
//                         {...field}
//                       />
//                       <FormMessage />
//                     </div>
//                   </FormItem>
//                 )}
//               />
//               <FormField
//                 control={form.control}
//                 name="password"
//                 render={({ field }) => (
//                   <FormItem>
//                     <div className="grid gap-2">
//                       <Label htmlFor="password">Mật khẩu</Label>
//                       <Input
//                         id="password"
//                         type="password"
//                         placeholder="••••••••"
//                         required
//                         className="border-2 border-gray-300 dark:border-gray-600"
//                         {...field}
//                       />
//                       <div className="text-sm text-gray-500 dark:text-gray-400 flex items-center gap-1">
//                         <Info size={16} />
//                         <span>
//                           Mật khẩu phải có ít nhất 8 ký tự và bao gồm chữ cái,
//                           số và ký tự đặc biệt.
//                         </span>
//                       </div>
//                       <FormMessage />
//                     </div>
//                   </FormItem>
//                 )}
//               />

//               <FormField
//                 control={form.control}
//                 name="phone"
//                 render={({ field }) => (
//                   <FormItem>
//                     <div className="grid gap-2">
//                       <Label htmlFor="phone">Số điện thoại</Label>
//                       <Input
//                         id="phone"
//                         type="tel"
//                         placeholder="1234567890"
//                         required
//                         className="border-2 border-gray-300 dark:border-gray-600"
//                         {...field}
//                         onChange={(e) => field.onChange(e.target.value)}
//                       />
//                       <FormMessage />
//                     </div>
//                   </FormItem>
//                 )}
//               />
//               <FormField
//                 control={form.control}
//                 name="address"
//                 render={({ field }) => (
//                   <FormItem>
//                     <div className="grid gap-2">
//                       <Label htmlFor="address">Địa chỉ</Label>
//                       <Input
//                         id="address"
//                         type="text"
//                         placeholder="123 Main St, Anytown, USA"
//                         required
//                         className="border-2 border-gray-300 dark:border-gray-600"
//                         {...field}
//                       />
//                       <FormMessage />
//                     </div>
//                   </FormItem>
//                 )}
//               />
//               <Button type="submit" className="w-full">
//                 Đăng ký
//               </Button>
//             </div>
//           </form>
//         </Form>
//       </DialogContent>
//     </Dialog>
//   );
// };

// export default RegisterDialog;
