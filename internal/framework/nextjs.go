package framework

import "strings"

// IsPublicEnvVar returns true if the key starts with NEXT_PUBLIC_,
// which means it will be exposed to the browser in Next.js.
func IsPublicEnvVar(key string) bool {
	return strings.HasPrefix(key, "NEXT_PUBLIC_")
}

// ClassifyNextJSEnv splits secrets into public (NEXT_PUBLIC_*) and private sets.
// Public variables are exposed to the browser at build time in Next.js.
func ClassifyNextJSEnv(secrets map[string]string) (public, private map[string]string) {
	public = make(map[string]string)
	private = make(map[string]string)
	for k, v := range secrets {
		if IsPublicEnvVar(k) {
			public[k] = v
		} else {
			private[k] = v
		}
	}
	return public, private
}
