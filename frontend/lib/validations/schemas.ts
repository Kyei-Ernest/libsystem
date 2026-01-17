import { z } from 'zod';

/**
 * Login Form Validation Schema
 */
export const loginSchema = z.object({
    email: z
        .string()
        .min(1, 'Email is required')
        .email('Invalid email address'),
    password: z
        .string()
        .min(1, 'Password is required')
        .min(8, 'Password must be at least 8 characters')
        .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
        .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
        .regex(/[0-9]/, 'Password must contain at least one number')
        .regex(/[^A-Za-z0-9]/, 'Password must contain at least one special character'),
});

/**
 * Register Form Validation Schema
 */
export const registerSchema = z.object({
    email: z
        .string()
        .min(1, 'Email is required')
        .email('Invalid email address'),
    username: z
        .string()
        .min(1, 'Username is required')
        .min(3, 'Username must be at least 3 characters')
        .max(50, 'Username must be less than 50 characters')
        .regex(/^[a-zA-Z0-9_-]+$/, 'Username can only contain letters, numbers, hyphens, and underscores'),
    firstName: z
        .string()
        .min(1, 'First name is required')
        .min(2, 'First name must be at least 2 characters')
        .max(50, 'First name must be less than 50 characters'),
    lastName: z
        .string()
        .min(1, 'Last name is required')
        .min(2, 'Last name must be at least 2 characters')
        .max(50, 'Last name must be less than 50 characters'),
    password: z
        .string()
        .min(1, 'Password is required')
        .min(8, 'Password must be at least 8 characters')
        .max(100, 'Password must be less than 100 characters')
        .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
        .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
        .regex(/[0-9]/, 'Password must contain at least one number')
        .regex(/[^A-Za-z0-9]/, 'Password must contain at least one special character (!@#$%^&* etc.)'),
    confirmPassword: z
        .string()
        .min(1, 'Please confirm your password'),
}).refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don't match",
    path: ['confirmPassword'],
});

/**
 * Collection Form Validation Schema
 */
export const collectionSchema = z.object({
    name: z
        .string()
        .min(1, 'Collection name is required')
        .min(3, 'Collection name must be at least 3 characters')
        .max(100, 'Collection name must be less than 100 characters'),
    description: z
        .string()
        .max(500, 'Description must be less than 500 characters')
        .optional(),
    is_public: z.boolean().default(false),
});

/**
 * Document Upload Validation Schema
 */
export const documentUploadSchema = z.object({
    title: z
        .string()
        .min(1, 'Title is required')
        .min(3, 'Title must be at least 3 characters')
        .max(200, 'Title must be less than 200 characters'),
    collection_id: z.string().min(1, 'Please select a collection'),
    file: z
        .instanceof(File)
        .refine((file) => file.size <= 104857600, 'File size must be less than 100MB')
        .refine(
            (file) => {
                const allowedTypes = ['pdf', 'docx', 'txt', 'jpg', 'jpeg', 'png', 'gif', 'html', 'htm'];
                const extension = file.name.split('.').pop()?.toLowerCase();
                return extension && allowedTypes.includes(extension);
            },
            'Invalid file type. Allowed: pdf, docx, txt, jpg, jpeg, png, gif, html, htm'
        ),
});

// Export types
export type LoginFormData = z.infer<typeof loginSchema>;
export type RegisterFormData = z.infer<typeof registerSchema>;
export type CollectionFormData = z.infer<typeof collectionSchema>;
export type DocumentUploadFormData = z.infer<typeof documentUploadSchema>;
