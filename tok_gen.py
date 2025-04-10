import jwt
import argparse
import datetime
import os
from typing import Dict, Any


def create_token(
    subject: str,
    name: str = None,
    expires_in: str = "24h",
    secret: str = None
) -> str:
    """
    Create a JWT token with the given claims.
    
    Args:
        subject: Subject identifier (usually user ID)
        name: Name of the user (optional)
        expires_in: Token expiration time (e.g. "24h", "7d")
        secret: Secret key for signing the token
        
    Returns:
        JWT token string
    """
    # Get secret from environment or use provided value
    secret = secret or os.environ.get("JWT_SECRET", "your-secret-key-here-replace-in-production")
    
    # Parse expiration time
    if expires_in.endswith('h'):
        expires_delta = datetime.timedelta(hours=int(expires_in[:-1]))
    elif expires_in.endswith('d'):
        expires_delta = datetime.timedelta(days=int(expires_in[:-1]))
    else:
        expires_delta = datetime.timedelta(hours=24)  # Default to 24 hours
    
    # Calculate timestamps
    issued_at = datetime.datetime.now()
    expires_at = issued_at + expires_delta
    
    # Prepare payload
    payload = {
        "sub": subject,
        "iat": int(issued_at.timestamp()),
        "exp": int(expires_at.timestamp()),
    }
    
    print(f"create token: {payload}")
    if name:
        payload["name"] = name
    
    # Create and sign the token
    token = jwt.encode(payload, secret, algorithm="HS256")
    return token

def verify_token(token: str, secret: str = None) -> Dict[str, Any]:
    """
    Verify and decode a JWT token.
    
    Args:
        token: JWT token string
        secret: Secret key used to sign the token
        
    Returns:
        Dictionary of claims from the token
    """
    secret = secret or os.environ.get("JWT_SECRET", "your-secret-key-here-replace-in-production")
    try:
        payload = jwt.decode(token, secret, algorithms=["HS256"])
        return payload
    except jwt.ExpiredSignatureError:
        raise ValueError("Token has expired")
    except jwt.InvalidTokenError:
        raise ValueError("Invalid token")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Generate JWT tokens for testing")
    parser.add_argument("--subject", "-s", required=True, help="Subject (user ID)")
    parser.add_argument("--name", "-n", help="User name")
    parser.add_argument("--expires", "-e", default="24h", help="Expiration time (e.g. 24h, 7d)")
    parser.add_argument("--secret", help="JWT secret key (default: from JWT_SECRET env var)")
    parser.add_argument("--verify", "-v", action="store_true", help="Verify the token instead of creating one")
    parser.add_argument("--token", "-t", help="Token to verify (required with --verify)")
    
    args = parser.parse_args()
    
    if args.verify:
        if not args.token:
            print("Error: --token is required when using --verify")
            exit(1)
        try:
            decoded = verify_token(args.token, args.secret)
            print("Token is valid. Payload:")
            for key, value in decoded.items():
                print(f"{key}: {value}")
        except ValueError as e:
            print(f"Error: {e}")
    else:
        token = create_token(args.subject, args.name, args.expires, args.secret)
        print(token)