// "use client";
// import React from "react";
// import { useForm } from "react-hook-form";
// import { zodResolver } from "@hookform/resolvers/zod";
// import { useRouter, useSearchParams } from "next/navigation";
// import { Button } from "@/components/ui/button";
// import {
//   Dialog,
//   DialogContent,
//   DialogHeader,
//   DialogTitle,
//   DialogTrigger
// } from "@/components/ui/dialog";
// import { Form, FormField, FormItem, FormMessage } from "@/components/ui/form";
// import { Input } from "@/components/ui/input";
// import { Label } from "@/components/ui/label";

// import { Info } from "lucide-react";

// import { handleErrorApi } from "@/lib/utils";
// import { LoginBody, LoginBodyType } from "../../../../auth/domain/auth.schema";
// import { useAuthStore } from "../../../../auth/controller/auth-controller";
// import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

// const LoginDialog: React.FC = () => {
//   const {
//     login,
//     isLoginDialogOpen,
//     openLoginDialog,
//     closeLoginDialog,
//     error,
//     clearError
//   } = useAuthStore();

//   const searchParams = useSearchParams();
//   const clearTokens = searchParams.get("clearTokens");

//   const form = useForm<LoginBodyType>({
//     resolver: zodResolver(LoginBody),
//     defaultValues: {
//       email: "",
//       password: ""
//     }
//   });

//   const router = useRouter();

//   const onSubmit = async (data: LoginBodyType) => {
//     try {
//       const loginResponse = await login(data);
//       const { role } = loginResponse.user;

//       // Role-based redirection
//       const roleRedirects = {
//         Admin: "/admin/dashboard",
//         Manager: "/manager/dashboard",
//         Employee: "/employee/dashboard",
//         Guest: "/dashboard"
//       };

//       const redirectPath = roleRedirects[role] || "/dashboard";
//       router.push(redirectPath);
//     } catch (error: any) {
//       console.error("Login error:", error);
//       // Handle login error (e.g., show error message to user)
//     }
//   };

//   return (
//     <Dialog
//       open={isLoginDialogOpen}
//       onOpenChange={(open) => {
//         if (open) {
//           openLoginDialog();
//         } else {
//           closeLoginDialog();
//           clearError(); // Clear any existing errors when closing the dialog
//         }
//       }}
//     >
//       <DialogTrigger asChild>
//         <Button onClick={openLoginDialog}>Đăng nhập</Button>
//       </DialogTrigger>
//       <DialogContent className="sm:max-w-[425px] bg-white dark:bg-gray-800 shadow-lg">
//         <DialogHeader>
//           <DialogTitle className="text-2xl">Đăng nhập</DialogTitle>
//           <p className="text-sm text-gray-500 dark:text-gray-400">
//             Nhập email và mật khẩu của bạn để đăng nhập vào hệ thống
//           </p>
//         </DialogHeader>

//         <Form {...form}>
//           <form
//             className="space-y-4 w-full"
//             noValidate
//             onSubmit={form.handleSubmit(onSubmit, (err) => {
//               console.log(err);
//             })}
//           >
//             <FormField
//               control={form.control}
//               name="email"
//               render={({ field }) => (
//                 <FormItem>
//                   <div className="grid gap-2">
//                     <Label htmlFor="email">Email</Label>
//                     <Input
//                       id="email"
//                       type="email"
//                       placeholder="m@example.com"
//                       required
//                       className="border-2 border-gray-300 dark:border-gray-600"
//                       {...field}
//                     />
//                     <FormMessage />
//                   </div>
//                 </FormItem>
//               )}
//             />
//             <FormField
//               control={form.control}
//               name="password"
//               render={({ field }) => (
//                 <FormItem>
//                   <div className="grid gap-2">
//                     <Label htmlFor="password">Password</Label>
//                     <Input
//                       placeholder="password"
//                       id="password"
//                       type="password"
//                       required
//                       className="border-2 border-gray-300 dark:border-gray-600"
//                       {...field}
//                     />
//                     <div className="text-sm text-gray-500 dark:text-gray-400 flex items-center gap-1">
//                       <Info size={16} />
//                       <span>
//                         Password should be at least 8 characters long and
//                         include a mix of letters, numbers, and symbols.
//                       </span>
//                     </div>
//                     <FormMessage />
//                   </div>
//                 </FormItem>
//               )}
//             />

//             {error && (
//               <Alert variant="destructive">
//                 <AlertTitle>Error</AlertTitle>
//                 <AlertDescription>{error}</AlertDescription>
//               </Alert>
//             )}
//             <Button type="submit" className="w-full">
//               Đăng nhập
//             </Button>
//             <Button variant="outline" className="w-full" type="button">
//               Đăng nhập bằng Google
//             </Button>
//           </form>
//         </Form>
//       </DialogContent>
//     </Dialog>
//   );
// };

// export default LoginDialog;
