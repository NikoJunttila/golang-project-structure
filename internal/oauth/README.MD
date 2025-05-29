# Google OAuth Client-Side Implementation

A concise guide for frontend developers to implement Google OAuth using a popup-based approach.

## Prerequisites

1. **Backend OAuth endpoints** (handled by your backend team):
   - `GET /google` - Redirects to Google's authorization URL
   - `POST /auth/google_callback` - Exchanges authorization code for tokens

2. **Browser Requirements**:
   - Popup windows must be allowed
   - Same-origin policy considerations

## Implementation

### 1. Main OAuth Function

```javascript
function googleLogin() {
    // Configure popup window
    const width = 500;
    const height = 600;
    const left = (window.innerWidth - width) / 2;
    const top = (window.innerHeight - height) / 2;
    
    // Open OAuth popup
    const popupWindow = window.open(
        'http://localhost:3000/google', // Your backend OAuth endpoint
        'Google Login',
        `width=${width},height=${height},top=${top},left=${left}`
    );
    
    if (!popupWindow) {
        console.error('Popup blocked by browser');
        return;
    }

    // Handle OAuth response
    handleOAuthResponse(popupWindow);
}
```

### 2. Handle OAuth Response

```javascript
function handleOAuthResponse(popupWindow) {
    // Listen for messages from popup
    const messageHandler = (event) => {
        // Security: Verify origin
        if (event.origin !== 'http://localhost:3000') {
            return;
        }

        console.log('OAuth success:', event.data);
        
        // Clean up
        window.removeEventListener('message', messageHandler);
        popupWindow.close();
        
        // Handle successful login
        onLoginSuccess(event.data);
    };

    window.addEventListener('message', messageHandler);

    // Monitor popup for closure/cancellation
    const checkClosed = setInterval(() => {
        if (popupWindow.closed) {
            clearInterval(checkClosed);
            window.removeEventListener('message', messageHandler);
            console.log('Login cancelled');
        }
    }, 1000);

    // Timeout after 5 minutes
    setTimeout(() => {
        if (!popupWindow.closed) {
            popupWindow.close();
            window.removeEventListener('message', messageHandler);
            clearInterval(checkClosed);
            console.error('Login timeout');
        }
    }, 5 * 60 * 1000);
}
```

### 3. Token Exchange (Fallback Method)

If your backend doesn't support postMessage, use URL monitoring:

```javascript
function monitorPopupUrl(popupWindow) {
    const interval = setInterval(() => {
        try {
            // Check if redirected to callback URL
            if (popupWindow.location.href.includes('/callback')) {
                const params = new URLSearchParams(popupWindow.location.search);
                const code = params.get('code');
                
                popupWindow.close();
                clearInterval(interval);
                
                // Exchange code for tokens
                exchangeCodeForTokens(code);
            }
        } catch (e) {
            // Ignore cross-origin errors
        }
    }, 1000);
}

async function exchangeCodeForTokens(code) {
    try {
        const response = await fetch('/auth/callback', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify({ code })
        });
        
        const data = await response.json();
        onLoginSuccess(data);
    } catch (error) {
        console.error('Token exchange failed:', error);
    }
}
```

## Complete Working Example

```javascript
class GoogleAuth {
    constructor(backendUrl = 'http://localhost:3000') {
        this.backendUrl = backendUrl;
        this.isLoading = false;
    }

    async login() {
        if (this.isLoading) return;
        
        this.isLoading = true;
        
        try {
            const popup = this.openPopup();
            if (!popup) {
                throw new Error('Popup blocked');
            }
            
            const result = await this.waitForAuth(popup);
            console.log('Login successful:', result);
            return result;
            
        } catch (error) {
            console.error('Login failed:', error);
            throw error;
        } finally {
            this.isLoading = false;
        }
    }

    openPopup() {
        const width = 500;
        const height = 600;
        const left = (window.innerWidth - width) / 2;
        const top = (window.innerHeight - height) / 2;
        
        return window.open(
            `${this.backendUrl}/google`,
            'Google Login',
            `width=${width},height=${height},top=${top},left=${left}`
        );
    }

    waitForAuth(popup) {
        return new Promise((resolve, reject) => {
            const messageHandler = (event) => {
                if (event.origin !== this.backendUrl) return;
                
                cleanup();
                resolve(event.data);
            };

            const checkClosed = setInterval(() => {
                if (popup.closed) {
                    cleanup();
                    reject(new Error('Login cancelled'));
                }
            }, 1000);

            const timeout = setTimeout(() => {
                cleanup();
                popup.close();
                reject(new Error('Login timeout'));
            }, 5 * 60 * 1000);

            const cleanup = () => {
                window.removeEventListener('message', messageHandler);
                clearInterval(checkClosed);
                clearTimeout(timeout);
            };

            window.addEventListener('message', messageHandler);
        });
    }
}

// Usage
const auth = new GoogleAuth();

document.getElementById('login-btn').addEventListener('click', async () => {
    try {
        const user = await auth.login();
        // Redirect or update UI
        window.location.href = '/dashboard';
    } catch (error) {
        alert('Login failed: ' + error.message);
    }
});
```

## HTML Integration

```html
<button id="google-login" onclick="handleGoogleLogin()">
    Sign in with Google
</button>

<script>
async function handleGoogleLogin() {
    const btn = document.getElementById('google-login');
    btn.disabled = true;
    btn.textContent = 'Signing in...';
    
    try {
        await googleLogin();
        location.reload(); // or redirect to dashboard
    } catch (error) {
        alert('Login failed. Please try again.');
    } finally {
        btn.disabled = false;
        btn.textContent = 'Sign in with Google';
    }
}
</script>
```

## React Example

```jsx
import { useState } from 'react';

function GoogleLoginButton() {
    const [isLoading, setIsLoading] = useState(false);

    const handleLogin = async () => {
        setIsLoading(true);
        try {
            await googleLogin();
            // Handle success
        } catch (error) {
            console.error('Login failed:', error);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <button onClick={handleLogin} disabled={isLoading}>
            {isLoading ? 'Signing in...' : 'Sign in with Google'}
        </button>
    );
}
```

## Common Issues & Solutions

**Popup Blocked**: Ensure the login function is called directly from user interaction (click event)

**Cross-Origin Errors**: Use postMessage instead of trying to access popup URL directly

**Cookies Not Set**: Make sure to include `credentials: 'include'` in fetch requests

**Login Timeout**: Increase timeout duration or add retry mechanism

## Security Notes

- Always validate `event.origin` in message handlers
- Use HTTPS in production
- Never store sensitive tokens in localStorage
- Let your backend handle token storage and validation

That's it! Your backend team handles the OAuth flow, you just need to manage the popup and handle the response.