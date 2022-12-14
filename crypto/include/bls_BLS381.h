/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

/**
 * @file bls_BLS381.h
 * @author Mike Scott
 * @date 28th Novemebr 2018
 * @brief BLS Header file
 *
 * Allows some user configuration
 * defines structures
 * declares functions
 *
 */

#ifndef BLS_BLS381_H
#define BLS_BLS381_H

#include "pair_BLS381.h"

/* Field size is assumed to be greater than or equal to group size */

#define BGS_BLS381 MODBYTES_384_58  /**< BLS Group Size */
#define BFS_BLS381 MODBYTES_384_58  /**< BLS Field Size */

#define BLS_OK             0   /**< Function completed without error */
#define BLS_FAIL	   41  /**< Invalid signature */
#define BLS_INVALID_G1     42  /**< Not a valid G1 point on the curve */
#define BLS_INVALID_G2     43  /**< Not a valid G2 point on the curve */

/* BLS API functions */

/**	@brief Generate Key Pair
 *
	@param RNG  Pointer to a cryptographically secure random number generator
	@param S    Private key. Generated externally if RNG set to NULL
	@param W    Public Key. W = S*G, where G is fixed generator
	@return     Zero for success or else an error code
 */
int BLS_BLS381_KEY_PAIR_GENERATE(csprng *RNG,octet* S,octet *W);

/**	@brief Calculate a signature
 *
	@param SIG  signature
	@param M    message to be signed
	@param S    Private key
	@return     Zero for success or else an error code
 */
int BLS_BLS381_SIGN(octet *SIG,octet *M,octet *S);

/**	@brief Verify a signature
 *
	@param SIG  signature
	@param M    message whose signature is to be verified.
	@param W    Public key
	@return     Zero for success or else an error code
 */
int BLS_BLS381_VERIFY(octet *SIG,octet *M,octet *W);

/**	@brief Add two members from the group G1
 *
	@param  R1  member of G1
	@param  R2  member of G1
	@param  R   member of G1. R = R1+R2
	@return     Zero for success or else an error code
 */
int BLS_BLS381_ADD_G1(octet *R1,octet *R2,octet *R);

/**	@brief Add two members from the group G2
 *
	@param  W1  member of G2
	@param  W2  member of G2
	@param  W   member of G2. W = W1+W2
	@return     Zero for success or else an error code
 */
int BLS_BLS381_ADD_G2(octet *W1,octet *W2,octet *W);

/**	@brief Use Shamir's secret sharing to distribute BLS secret keys
 *
	@param  k   Threshold
	@param  n   Number of shares
        @param  RNG Pointer to a cryptographically secure random number generator
	@param  X   X values
	@param  Y   Y values. Valid BLS secret keys
	@param  SKI Input secret key to be shared. Ignored if set to NULL
	@param  SKO Secret key that is shared
	@return     Zero for success or else an error code
 */
int BLS_BLS381_MAKE_SHARES(int k, int n, csprng *RNG, octet* X, octet* Y, octet* SKI, octet* SKO);

/**	@brief Use Shamir's secret sharing to recover a BLS secret key
 *
	@param  k   Threshold
	@param  X   X values
	@param  Y   Y values. Valid BLS secret keys
	@param  SK  Secret key that is recovered
	@return     Zero for success or else an error code
 */
int BLS_BLS381_RECOVER_SECRET(int k, octet* X, octet* Y, octet* SK);

/**	@brief Use Shamir's secret sharing to recover a BLS signature
 *
	@param  k   Threshold
	@param  X   X values
	@param  Y   Y values. Valid BLS signatures
	@param  SIG Signature that is recovered
	@return     Zero for success or else an error code
 */
int BLS_BLS381_RECOVER_SIGNATURE(int k, octet* X, octet* Y, octet* SIG);

#endif

