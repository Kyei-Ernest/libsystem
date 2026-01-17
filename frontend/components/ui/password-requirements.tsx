export function PasswordRequirements() {
    return (
        <div className="rounded-md bg-blue-50 p-3 text-sm">
            <p className="font-medium text-blue-900 mb-1">Password must contain:</p>
            <ul className="list-disc list-inside space-y-1 text-blue-700">
                <li>At least 8 characters</li>
                <li>One uppercase letter (A-Z)</li>
                <li>One lowercase letter (a-z)</li>
                <li>One number (0-9)</li>
                <li>One special character (!@#$%^&* etc.)</li>
            </ul>
        </div>
    );
}
