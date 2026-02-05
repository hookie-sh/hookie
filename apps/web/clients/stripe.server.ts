// FAKE CARDS FOR TESTING PURPOSES => https://docs.stripe.com/testing#cards

import assert from "node:assert";
import Stripe from "stripe";

assert(process.env.STRIPE_SECRET_KEY, "STRIPE_SECRET_KEY is not set");
assert(process.env.NEXT_PUBLIC_APP_URL, "NEXT_PUBLIC_APP_URL is not set");

export const stripe = new Stripe(process.env.STRIPE_SECRET_KEY);
