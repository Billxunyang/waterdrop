/*
 *
 * Copyright 2020 waterdrop authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package http

import (
	"context"

	"google.golang.org/grpc/metadata"

	tracer "github.com/UnderTreeTech/waterdrop/pkg/trace"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go/ext"
)

func (s *Server) trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, opt := tracer.HeaderExtractor(c.Request.Context(), c.Request.Header)
		span, ctx := tracer.StartSpanFromContext(ctx, c.Request.Method+" "+c.Request.URL.Path, opt)
		ext.Component.Set(span, "http")
		ext.SpanKind.Set(span, ext.SpanKindRPCServerEnum)
		ext.HTTPMethod.Set(span, c.Request.Method)
		ext.HTTPUrl.Set(span, c.Request.URL.Path)
		ext.PeerHostIPv4.SetString(span, c.ClientIP())

		// adjust request timeout
		timeout := s.config.Timeout
		reqTimeout := getTimeout(c.Request)
		if reqTimeout > 0 && timeout > reqTimeout {
			timeout = reqTimeout
		}

		ctx = metadata.NewIncomingContext(ctx, metadata.MD(c.Request.Header))
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer func() {
			span.Finish()
			cancel()
		}()

		c.Request = c.Request.WithContext(ctx)
		c.Writer.Header().Set(_httpHeaderTraceId, tracer.TraceID(ctx))

		c.Next()
	}
}
