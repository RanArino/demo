import { SignUp } from "@clerk/nextjs";

export default function Page() {
  return (
    <SignUp 
      forceRedirectUrl="/profile/setup"
      appearance={{
        elements: {
          rootBox: {
            width: "100%",
          },
          card: {
            width: "100%",
            boxShadow: "0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)",
            borderRadius: "8px",
          },
        }
      }}
    />
  );
}
