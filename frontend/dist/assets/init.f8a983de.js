import{E as g,U as gt,T as pe,G as Me,b as Ge,M as F,c as ne,d as z,B as Ae,f as P,g as B,R as ie,w as Q,h as ze,i as L,v as De,j as E,k as S,D as K,n as me,l as J,m as A,o as Y,p as xt,q as ke,s as ge,t as X,u as _t,x as se,y as Z,V,S as ae,z as ee,A as Oe,F as Ie,H as Ee,I as Ve,J as vt,K as bt,L as yt,N as Tt,O as wt,Q as Ct,W as Pt,X as St,Y as te,Z as Ft,_ as xe,e as T,$ as Rt}from"./index.cc4c8d50.js";import{F as Ut,S as H,c as N,l as Bt,a as Mt,B as We}from"./colorToUniform.718bd2cf.js";class Le{static init(e){Object.defineProperty(this,"resizeTo",{configurable:!0,set(t){globalThis.removeEventListener("resize",this.queueResize),this._resizeTo=t,t&&(globalThis.addEventListener("resize",this.queueResize),this.resize())},get(){return this._resizeTo}}),this.queueResize=()=>{!this._resizeTo||(this._cancelResize(),this._resizeId=requestAnimationFrame(()=>this.resize()))},this._cancelResize=()=>{this._resizeId&&(cancelAnimationFrame(this._resizeId),this._resizeId=null)},this.resize=()=>{if(!this._resizeTo)return;this._cancelResize();let t,r;if(this._resizeTo===globalThis.window)t=globalThis.innerWidth,r=globalThis.innerHeight;else{const{clientWidth:n,clientHeight:i}=this._resizeTo;t=n,r=i}this.renderer.resize(t,r),this.render()},this._resizeId=null,this._resizeTo=null,this.resizeTo=e.resizeTo||null}static destroy(){globalThis.removeEventListener("resize",this.queueResize),this._cancelResize(),this._cancelResize=null,this.queueResize=null,this.resizeTo=null,this.resize=null}}Le.extension=g.Application;class Ye{static init(e){e=Object.assign({autoStart:!0,sharedTicker:!1},e),Object.defineProperty(this,"ticker",{configurable:!0,set(t){this._ticker&&this._ticker.remove(this.render,this),this._ticker=t,t&&t.add(this.render,this,gt.LOW)},get(){return this._ticker}}),this.stop=()=>{this._ticker.stop()},this.start=()=>{this._ticker.start()},this._ticker=null,this.ticker=e.sharedTicker?pe.shared:new pe,e.autoStart&&this.start()}static destroy(){if(this._ticker){const e=this._ticker;this.ticker=null,e.destroy()}}}Ye.extension=g.Application;var Gt=`in vec2 aPosition;
out vec2 vTextureCoord;

uniform vec4 uInputSize;
uniform vec4 uOutputFrame;
uniform vec4 uOutputTexture;

vec4 filterVertexPosition( void )
{
    vec2 position = aPosition * uOutputFrame.zw + uOutputFrame.xy;
    
    position.x = position.x * (2.0 / uOutputTexture.x) - 1.0;
    position.y = position.y * (2.0*uOutputTexture.z / uOutputTexture.y) - uOutputTexture.z;

    return vec4(position, 0.0, 1.0);
}

vec2 filterTextureCoord( void )
{
    return aPosition * (uOutputFrame.zw * uInputSize.zw);
}

void main(void)
{
    gl_Position = filterVertexPosition();
    vTextureCoord = filterTextureCoord();
}
`,At=`in vec2 vTextureCoord;
out vec4 finalColor;
uniform sampler2D uTexture;
void main() {
    finalColor = texture(uTexture, vTextureCoord);
}
`,_e=`struct GlobalFilterUniforms {
  uInputSize: vec4<f32>,
  uInputPixel: vec4<f32>,
  uInputClamp: vec4<f32>,
  uOutputFrame: vec4<f32>,
  uGlobalFrame: vec4<f32>,
  uOutputTexture: vec4<f32>,
};

@group(0) @binding(0) var <uniform> gfu: GlobalFilterUniforms;
@group(0) @binding(1) var uTexture: texture_2d<f32>;
@group(0) @binding(2) var uSampler: sampler;

struct VSOutput {
  @builtin(position) position: vec4<f32>,
  @location(0) uv: vec2<f32>
};

fn filterVertexPosition(aPosition: vec2<f32>) -> vec4<f32>
{
    var position = aPosition * gfu.uOutputFrame.zw + gfu.uOutputFrame.xy;

    position.x = position.x * (2.0 / gfu.uOutputTexture.x) - 1.0;
    position.y = position.y * (2.0 * gfu.uOutputTexture.z / gfu.uOutputTexture.y) - gfu.uOutputTexture.z;

    return vec4(position, 0.0, 1.0);
}

fn filterTextureCoord(aPosition: vec2<f32>) -> vec2<f32>
{
    return aPosition * (gfu.uOutputFrame.zw * gfu.uInputSize.zw);
}

@vertex
fn mainVertex(
  @location(0) aPosition: vec2<f32>,
) -> VSOutput {
  return VSOutput(
   filterVertexPosition(aPosition),
   filterTextureCoord(aPosition)
  );
}

@fragment
fn mainFragment(
  @location(0) uv: vec2<f32>,
) -> @location(0) vec4<f32> {
    return textureSample(uTexture, uSampler, uv);
}
`;class zt extends Ut{constructor(){const e=Me.from({vertex:{source:_e,entryPoint:"mainVertex"},fragment:{source:_e,entryPoint:"mainFragment"},name:"passthrough-filter"}),t=Ge.from({vertex:Gt,fragment:At,name:"passthrough-filter"});super({gpuProgram:e,glProgram:t})}}class Xe{constructor(e){this._renderer=e}push(e,t,r){this._renderer.renderPipes.batch.break(r),r.add({renderPipeId:"filter",canBundle:!1,action:"pushFilter",container:t,filterEffect:e})}pop(e,t,r){this._renderer.renderPipes.batch.break(r),r.add({renderPipeId:"filter",action:"popFilter",canBundle:!1})}execute(e){e.action==="pushFilter"?this._renderer.filter.push(e):e.action==="popFilter"&&this._renderer.filter.pop()}destroy(){this._renderer=null}}Xe.extension={type:[g.WebGLPipes,g.WebGPUPipes,g.CanvasPipes],name:"filter"};const ve=new F;function Dt(o,e){var r;e.clear();const t=e.matrix;for(let n=0;n<o.length;n++){const i=o[n];if(i.globalDisplayStatus<7)continue;const s=(r=i.renderGroup)!=null?r:i.parentRenderGroup;s!=null&&s.isCachedAsTexture?e.matrix=ve.copyFrom(s.textureOffsetInverseTransform).append(i.worldTransform):s!=null&&s._parentCacheAsTextureRenderGroup?e.matrix=ve.copyFrom(s._parentCacheAsTextureRenderGroup.inverseWorldTransform).append(i.groupTransform):e.matrix=i.worldTransform,e.addBounds(i.bounds)}return e.matrix=t,e}const kt=new ne({attributes:{aPosition:{buffer:new Float32Array([0,0,1,0,1,1,0,1]),format:"float32x2",stride:2*4,offset:0}},indexBuffer:new Uint32Array([0,1,2,0,2,3])});class Ot{constructor(){this.skip=!1,this.inputTexture=null,this.backTexture=null,this.filters=null,this.bounds=new ze,this.container=null,this.blendRequired=!1,this.outputRenderSurface=null,this.globalFrame={x:0,y:0,width:0,height:0},this.firstEnabledIndex=-1,this.lastEnabledIndex=-1}}class He{constructor(e){this._filterStackIndex=0,this._filterStack=[],this._filterGlobalUniforms=new z({uInputSize:{value:new Float32Array(4),type:"vec4<f32>"},uInputPixel:{value:new Float32Array(4),type:"vec4<f32>"},uInputClamp:{value:new Float32Array(4),type:"vec4<f32>"},uOutputFrame:{value:new Float32Array(4),type:"vec4<f32>"},uGlobalFrame:{value:new Float32Array(4),type:"vec4<f32>"},uOutputTexture:{value:new Float32Array(4),type:"vec4<f32>"}}),this._globalFilterBindGroup=new Ae({}),this.renderer=e}get activeBackTexture(){var e;return(e=this._activeFilterData)==null?void 0:e.backTexture}push(e){const t=this.renderer,r=e.filterEffect.filters,n=this._pushFilterData();n.skip=!1,n.filters=r,n.container=e.container,n.outputRenderSurface=t.renderTarget.renderSurface;const i=t.renderTarget.renderTarget.colorTexture.source,s=i.resolution,a=i.antialias;if(r.every(f=>!f.enabled)){n.skip=!0;return}const l=n.bounds;if(this._calculateFilterArea(e,l),this._calculateFilterBounds(n,t.renderTarget.rootViewPort,a,s,1),n.skip)return;const u=this._getPreviousFilterData(),c=this._findFilterResolution(s);let h=0,d=0;u&&(h=u.bounds.minX,d=u.bounds.minY),this._calculateGlobalFrame(n,h,d,c,i.width,i.height),this._setupFilterTextures(n,l,t,u)}generateFilteredTexture({texture:e,filters:t}){const r=this._pushFilterData();this._activeFilterData=r,r.skip=!1,r.filters=t;const n=e.source,i=n.resolution,s=n.antialias;if(t.every(f=>!f.enabled))return r.skip=!0,e;const a=r.bounds;if(a.addRect(e.frame),this._calculateFilterBounds(r,a.rectangle,s,i,0),r.skip)return e;const l=i,u=0,c=0;this._calculateGlobalFrame(r,u,c,l,n.width,n.height),r.outputRenderSurface=P.getOptimalTexture(a.width,a.height,r.resolution,r.antialias),r.backTexture=B.EMPTY,r.inputTexture=e,this.renderer.renderTarget.finishRenderPass(),this._applyFiltersToTexture(r,!0);const d=r.outputRenderSurface;return d.source.alphaMode="premultiplied-alpha",d}pop(){const e=this.renderer,t=this._popFilterData();t.skip||(e.globalUniforms.pop(),e.renderTarget.finishRenderPass(),this._activeFilterData=t,this._applyFiltersToTexture(t,!1),t.blendRequired&&P.returnTexture(t.backTexture),P.returnTexture(t.inputTexture))}getBackTexture(e,t,r){const n=e.colorTexture.source._resolution,i=P.getOptimalTexture(t.width,t.height,n,!1);let s=t.minX,a=t.minY;r&&(s-=r.minX,a-=r.minY),s=Math.floor(s*n),a=Math.floor(a*n);const l=Math.ceil(t.width*n),u=Math.ceil(t.height*n);return this.renderer.renderTarget.copyToTexture(e,i,{x:s,y:a},{width:l,height:u},{x:0,y:0}),i}applyFilter(e,t,r,n){const i=this.renderer,s=this._activeFilterData,l=s.outputRenderSurface===r,u=i.renderTarget.rootRenderTarget.colorTexture.source._resolution,c=this._findFilterResolution(u);let h=0,d=0;if(l){const x=this._findPreviousFilterOffset();h=x.x,d=x.y}this._updateFilterUniforms(t,r,s,h,d,c,l,n);const f=e.enabled?e:this._getPassthroughFilter();this._setupBindGroupsAndRender(f,t,i)}calculateSpriteMatrix(e,t){const r=this._activeFilterData,n=e.set(r.inputTexture._source.width,0,0,r.inputTexture._source.height,r.bounds.minX,r.bounds.minY),i=t.worldTransform.copyTo(F.shared),s=t.renderGroup||t.parentRenderGroup;return s&&s.cacheToLocalTransform&&i.prepend(s.cacheToLocalTransform),i.invert(),n.prepend(i),n.scale(1/t.texture.orig.width,1/t.texture.orig.height),n.translate(t.anchor.x,t.anchor.y),n}destroy(){var e;(e=this._passthroughFilter)==null||e.destroy(!0),this._passthroughFilter=null}_getPassthroughFilter(){var e;return(e=this._passthroughFilter)!=null||(this._passthroughFilter=new zt),this._passthroughFilter}_setupBindGroupsAndRender(e,t,r){if(r.renderPipes.uniformBatch){const n=r.renderPipes.uniformBatch.getUboResource(this._filterGlobalUniforms);this._globalFilterBindGroup.setResource(n,0)}else this._globalFilterBindGroup.setResource(this._filterGlobalUniforms,0);this._globalFilterBindGroup.setResource(t.source,1),this._globalFilterBindGroup.setResource(t.source.style,2),e.groups[0]=this._globalFilterBindGroup,r.encoder.draw({geometry:kt,shader:e,state:e._state,topology:"triangle-list"}),r.type===ie.WEBGL&&r.renderTarget.finishRenderPass()}_setupFilterTextures(e,t,r,n){if(e.backTexture=B.EMPTY,e.inputTexture=P.getOptimalTexture(t.width,t.height,e.resolution,e.antialias),e.blendRequired){r.renderTarget.finishRenderPass();const i=r.renderTarget.getRenderTarget(e.outputRenderSurface);e.backTexture=this.getBackTexture(i,t,n==null?void 0:n.bounds)}r.renderTarget.bind(e.inputTexture,!0),r.globalUniforms.push({offset:t})}_calculateGlobalFrame(e,t,r,n,i,s){const a=e.globalFrame;a.x=t*n,a.y=r*n,a.width=i*n,a.height=s*n}_updateFilterUniforms(e,t,r,n,i,s,a,l){const u=this._filterGlobalUniforms.uniforms,c=u.uOutputFrame,h=u.uInputSize,d=u.uInputPixel,f=u.uInputClamp,x=u.uGlobalFrame,p=u.uOutputTexture;a?(c[0]=r.bounds.minX-n,c[1]=r.bounds.minY-i):(c[0]=0,c[1]=0),c[2]=e.frame.width,c[3]=e.frame.height,h[0]=e.source.width,h[1]=e.source.height,h[2]=1/h[0],h[3]=1/h[1],d[0]=e.source.pixelWidth,d[1]=e.source.pixelHeight,d[2]=1/d[0],d[3]=1/d[1],f[0]=.5*d[2],f[1]=.5*d[3],f[2]=e.frame.width*h[2]-.5*d[2],f[3]=e.frame.height*h[3]-.5*d[3];const m=this.renderer.renderTarget.rootRenderTarget.colorTexture;x[0]=n*s,x[1]=i*s,x[2]=m.source.width*s,x[3]=m.source.height*s,t instanceof B&&(t.source.resource=null);const _=this.renderer.renderTarget.getRenderTarget(t);this.renderer.renderTarget.bind(t,!!l),t instanceof B?(p[0]=t.frame.width,p[1]=t.frame.height):(p[0]=_.width,p[1]=_.height),p[2]=_.isRoot?-1:1,this._filterGlobalUniforms.update()}_findFilterResolution(e){let t=this._filterStackIndex-1;for(;t>0&&this._filterStack[t].skip;)--t;return t>0&&this._filterStack[t].inputTexture?this._filterStack[t].inputTexture.source._resolution:e}_findPreviousFilterOffset(){let e=0,t=0,r=this._filterStackIndex;for(;r>0;){r--;const n=this._filterStack[r];if(!n.skip){e=n.bounds.minX,t=n.bounds.minY;break}}return{x:e,y:t}}_calculateFilterArea(e,t){if(e.renderables?Dt(e.renderables,t):e.filterEffect.filterArea?(t.clear(),t.addRect(e.filterEffect.filterArea),t.applyMatrix(e.container.worldTransform)):e.container.getFastGlobalBounds(!0,t),e.container){const n=(e.container.renderGroup||e.container.parentRenderGroup).cacheToLocalTransform;n&&t.applyMatrix(n)}}_applyFiltersToTexture(e,t){const r=e.inputTexture,n=e.bounds,i=e.filters,s=e.firstEnabledIndex,a=e.lastEnabledIndex;if(this._globalFilterBindGroup.setResource(r.source.style,2),this._globalFilterBindGroup.setResource(e.backTexture.source,3),s===a)i[s].apply(this,r,e.outputRenderSurface,t);else{let l=e.inputTexture;const u=P.getOptimalTexture(n.width,n.height,l.source._resolution,!1);let c=u;for(let h=s;h<a;h++){const d=i[h];if(!d.enabled)continue;d.apply(this,l,c,!0);const f=l;l=c,c=f}i[a].apply(this,l,e.outputRenderSurface,t),P.returnTexture(u)}}_calculateFilterBounds(e,t,r,n,i){var _,C;const s=this.renderer,a=e.bounds,l=e.filters;let u=1/0,c=0,h=!0,d=!1,f=!1,x=!0,p=-1,m=-1;for(let b=0;b<l.length;b++){const v=l[b];if(!v.enabled)continue;if(p===-1&&(p=b),m=b,u=Math.min(u,v.resolution==="inherit"?n:v.resolution),c+=v.padding,v.antialias==="off"?h=!1:v.antialias==="inherit"&&h&&(h=r),v.clipToViewport||(x=!1),!!!(v.compatibleRenderers&s.type)){f=!1;break}if(v.blendRequired&&!((C=(_=s.backBuffer)==null?void 0:_.useBackBuffer)==null||C)){Q("Blend filter requires backBuffer on WebGL renderer to be enabled. Set `useBackBuffer: true` in the renderer options."),f=!1;break}f=!0,d||(d=v.blendRequired)}if(!f){e.skip=!0;return}if(x&&a.fitBounds(0,t.width/n,0,t.height/n),a.scale(u).ceil().scale(1/u).pad((c|0)*i),!a.isPositive){e.skip=!0;return}e.antialias=h,e.resolution=u,e.blendRequired=d,e.firstEnabledIndex=p,e.lastEnabledIndex=m}_popFilterData(){return this._filterStackIndex--,this._filterStack[this._filterStackIndex]}_getPreviousFilterData(){let e,t=this._filterStackIndex-1;for(;t>0&&(t--,e=this._filterStack[t],!!e.skip););return e}_pushFilterData(){let e=this._filterStack[this._filterStackIndex];return e||(e=this._filterStack[this._filterStackIndex]=new Ot),this._filterStackIndex++,e}}He.extension={type:[g.WebGLSystem,g.WebGPUSystem],name:"filter"};const Ke=class Ne extends ne{constructor(...e){var c;let t=(c=e[0])!=null?c:{};t instanceof Float32Array&&(L(De,"use new MeshGeometry({ positions, uvs, indices }) instead"),t={positions:t,uvs:e[1],indices:e[2]}),t={...Ne.defaultOptions,...t};const r=t.positions||new Float32Array([0,0,1,0,1,1,0,1]);let n=t.uvs;n||(t.positions?n=new Float32Array(r.length):n=new Float32Array([0,0,1,0,1,1,0,1]));const i=t.indices||new Uint32Array([0,1,2,0,2,3]),s=t.shrinkBuffersToFit,a=new E({data:r,label:"attribute-mesh-positions",shrinkToFit:s,usage:S.VERTEX|S.COPY_DST}),l=new E({data:n,label:"attribute-mesh-uvs",shrinkToFit:s,usage:S.VERTEX|S.COPY_DST}),u=new E({data:i,label:"index-mesh-buffer",shrinkToFit:s,usage:S.INDEX|S.COPY_DST});super({attributes:{aPosition:{buffer:a,format:"float32x2",stride:2*4,offset:0},aUV:{buffer:l,format:"float32x2",stride:2*4,offset:0}},indexBuffer:u,topology:t.topology}),this.batchMode="auto"}get positions(){return this.attributes.aPosition.buffer.data}set positions(e){this.attributes.aPosition.buffer.data=e}get uvs(){return this.attributes.aUV.buffer.data}set uvs(e){this.attributes.aUV.buffer.data=e}get indices(){return this.indexBuffer.data}set indices(e){this.indexBuffer.data=e}};Ke.defaultOptions={topology:"triangle-list",shrinkBuffersToFit:!1};let oe=Ke,M=null,R=null;function It(o,e){M||(M=K.get().createCanvas(256,128),R=M.getContext("2d",{willReadFrequently:!0}),R.globalCompositeOperation="copy",R.globalAlpha=1),(M.width<o||M.height<e)&&(M.width=me(o),M.height=me(e))}function be(o,e,t){for(let r=0,n=4*t*e;r<e;++r,n+=4)if(o[n+3]!==0)return!1;return!0}function ye(o,e,t,r,n){const i=4*e;for(let s=r,a=r*i+4*t;s<=n;++s,a+=i)if(o[a+3]!==0)return!1;return!0}function Et(...o){var f,x,p;let e=o[0];e.canvas||(e={canvas:o[0],resolution:o[1]});const{canvas:t}=e,r=Math.min((f=e.resolution)!=null?f:1,1),n=(x=e.width)!=null?x:t.width,i=(p=e.height)!=null?p:t.height;let s=e.output;if(It(n,i),!R)throw new TypeError("Failed to get canvas 2D context");R.drawImage(t,0,0,n,i,0,0,n*r,i*r);const l=R.getImageData(0,0,n,i).data;let u=0,c=0,h=n-1,d=i-1;for(;c<i&&be(l,n,c);)++c;if(c===i)return J.EMPTY;for(;be(l,n,d);)--d;for(;ye(l,n,u,c,d);)++u;for(;ye(l,n,h,c,d);)--h;return++h,++d,R.globalCompositeOperation="source-over",R.strokeRect(u,c,h-u,d-c),R.globalCompositeOperation="copy",s!=null||(s=new J),s.set(u/r,c/r,(h-u)/r,(d-c)/r),s}const Te=new J;class Vt{getCanvasAndContext(e){const{text:t,style:r,resolution:n=1}=e,i=r._getFinalPadding(),s=A.measureText(t||" ",r),a=Math.ceil(Math.ceil(Math.max(1,s.width)+i*2)*n),l=Math.ceil(Math.ceil(Math.max(1,s.height)+i*2)*n),u=Y.getOptimalCanvasAndContext(a,l);this._renderTextToCanvas(t,r,i,n,u);const c=r.trim?Et({canvas:u.canvas,width:a,height:l,resolution:1,output:Te}):Te.set(0,0,a,l);return{canvasAndContext:u,frame:c}}returnCanvasAndContext(e){Y.returnCanvasAndContext(e)}_renderTextToCanvas(e,t,r,n,i){var b,v,w,G,le;const{canvas:s,context:a}=i,l=xt(t),u=A.measureText(e||" ",t),c=u.lines,h=u.lineHeight,d=u.lineWidths,f=u.maxLineWidth,x=u.fontProperties,p=s.height;if(a.resetTransform(),a.scale(n,n),a.textBaseline=t.textBaseline,(b=t._stroke)!=null&&b.width){const U=t._stroke;a.lineWidth=U.width,a.miterLimit=U.miterLimit,a.lineJoin=U.join,a.lineCap=U.cap}a.font=l;let m,_;const C=t.dropShadow?2:1;for(let U=0;U<C;++U){const ce=t.dropShadow&&U===0,j=ce?Math.ceil(Math.max(1,p)+r*2):0,ht=j*n;if(ce){a.fillStyle="black",a.strokeStyle="black";const y=t.dropShadow,ft=y.color,pt=y.alpha;a.shadowColor=ke.shared.setValue(ft).setAlpha(pt).toRgbaString();const mt=y.blur*n,fe=y.distance*n;a.shadowBlur=mt,a.shadowOffsetX=Math.cos(y.angle)*fe,a.shadowOffsetY=Math.sin(y.angle)*fe+ht}else{if(a.fillStyle=t._fill?ge(t._fill,a,u,r*2):null,(v=t._stroke)!=null&&v.width){const y=t._stroke.width*.5+r*2;a.strokeStyle=ge(t._stroke,a,u,y)}a.shadowColor="black"}let de=(h-x.fontSize)/2;h-x.fontSize<0&&(de=0);const he=(G=(w=t._stroke)==null?void 0:w.width)!=null?G:0;for(let y=0;y<c.length;y++)m=he/2,_=he/2+y*h+x.ascent+de,t.align==="right"?m+=f-d[y]:t.align==="center"&&(m+=(f-d[y])/2),(le=t._stroke)!=null&&le.width&&this._drawLetterSpacing(c[y],t,i,m+r,_+r-j,!0),t._fill!==void 0&&this._drawLetterSpacing(c[y],t,i,m+r,_+r-j)}}_drawLetterSpacing(e,t,r,n,i,s=!1){const{context:a}=r,l=t.letterSpacing;let u=!1;if(A.experimentalLetterSpacingSupported&&(A.experimentalLetterSpacing?(a.letterSpacing=`${l}px`,a.textLetterSpacing=`${l}px`,u=!0):(a.letterSpacing="0px",a.textLetterSpacing="0px")),l===0||u){s?a.strokeText(e,n,i):a.fillText(e,n,i);return}let c=n;const h=A.graphemeSegmenter(e);let d=a.measureText(e).width,f=0;for(let x=0;x<h.length;++x){const p=h[x];s?a.strokeText(p,c,i):a.fillText(p,c,i);let m="";for(let _=x+1;_<h.length;++_)m+=h[_];f=a.measureText(m).width,c+=d-f+l,d=f}}}const $=new Vt,we="http://www.w3.org/2000/svg",Ce="http://www.w3.org/1999/xhtml";class je{constructor(){this.svgRoot=document.createElementNS(we,"svg"),this.foreignObject=document.createElementNS(we,"foreignObject"),this.domElement=document.createElementNS(Ce,"div"),this.styleElement=document.createElementNS(Ce,"style");const{foreignObject:e,svgRoot:t,styleElement:r,domElement:n}=this;e.setAttribute("width","10000"),e.setAttribute("height","10000"),e.style.overflow="hidden",t.appendChild(e),e.appendChild(r),e.appendChild(n),this.image=K.get().createImage()}destroy(){this.svgRoot.remove(),this.foreignObject.remove(),this.styleElement.remove(),this.domElement.remove(),this.image.src="",this.image.remove(),this.svgRoot=null,this.foreignObject=null,this.styleElement=null,this.domElement=null,this.image=null,this.canvasAndContext=null}}let Pe;function Wt(o,e,t,r){r||(r=Pe||(Pe=new je));const{domElement:n,styleElement:i,svgRoot:s}=r;n.innerHTML=`<style>${e.cssStyle};</style><div style='padding:0'>${o}</div>`,n.setAttribute("style","transform-origin: top left; display: inline-block"),t&&(i.textContent=t),document.body.appendChild(s);const a=n.getBoundingClientRect();s.remove();const l=e.padding*2;return{width:a.width-l,height:a.height-l}}class Lt{constructor(){this.batches=[],this.batched=!1}destroy(){this.batches.forEach(e=>{X.return(e)}),this.batches.length=0}}class $e{constructor(e,t){this.state=H.for2d(),this.renderer=e,this._adaptor=t,this.renderer.runners.contextChange.add(this)}contextChange(){this._adaptor.contextChange(this.renderer)}validateRenderable(e){const t=e.context,r=!!e._gpuData,n=this.renderer.graphicsContext.updateGpuContext(t);return!!(n.isBatchable||r!==n.isBatchable)}addRenderable(e,t){const r=this.renderer.graphicsContext.updateGpuContext(e.context);e.didViewUpdate&&this._rebuild(e),r.isBatchable?this._addToBatcher(e,t):(this.renderer.renderPipes.batch.break(t),t.add(e))}updateRenderable(e){const r=this._getGpuDataForRenderable(e).batches;for(let n=0;n<r.length;n++){const i=r[n];i._batcher.updateElement(i)}}execute(e){if(!e.isRenderable)return;const t=this.renderer,r=e.context;if(!t.graphicsContext.getGpuContext(r).batches.length)return;const i=r.customShader||this._adaptor.shader;this.state.blendMode=e.groupBlendMode;const s=i.resources.localUniforms.uniforms;s.uTransformMatrix=e.groupTransform,s.uRound=t._roundPixels|e._roundPixels,N(e.groupColorAlpha,s.uColor,0),this._adaptor.execute(this,e)}_rebuild(e){const t=this._getGpuDataForRenderable(e),r=this.renderer.graphicsContext.updateGpuContext(e.context);t.destroy(),r.isBatchable&&this._updateBatchesForRenderable(e,t)}_addToBatcher(e,t){const r=this.renderer.renderPipes.batch,n=this._getGpuDataForRenderable(e).batches;for(let i=0;i<n.length;i++){const s=n[i];r.addToBatch(s,t)}}_getGpuDataForRenderable(e){return e._gpuData[this.renderer.uid]||this._initGpuDataForRenderable(e)}_initGpuDataForRenderable(e){const t=new Lt;return e._gpuData[this.renderer.uid]=t,t}_updateBatchesForRenderable(e,t){const r=e.context,n=this.renderer.graphicsContext.getGpuContext(r),i=this.renderer._roundPixels|e._roundPixels;t.batches=n.batches.map(s=>{const a=X.get(_t);return s.copyTo(a),a.renderable=e,a.roundPixels=i,a})}destroy(){this.renderer=null,this._adaptor.destroy(),this._adaptor=null,this.state=null}}$e.extension={type:[g.WebGLPipes,g.WebGPUPipes,g.CanvasPipes],name:"graphics"};const qe=class Qe extends oe{constructor(...e){var r;super({});let t=(r=e[0])!=null?r:{};typeof t=="number"&&(L(De,"PlaneGeometry constructor changed please use { width, height, verticesX, verticesY } instead"),t={width:t,height:e[1],verticesX:e[2],verticesY:e[3]}),this.build(t)}build(e){var h,d,f,x;e={...Qe.defaultOptions,...e},this.verticesX=(h=this.verticesX)!=null?h:e.verticesX,this.verticesY=(d=this.verticesY)!=null?d:e.verticesY,this.width=(f=this.width)!=null?f:e.width,this.height=(x=this.height)!=null?x:e.height;const t=this.verticesX*this.verticesY,r=[],n=[],i=[],s=this.verticesX-1,a=this.verticesY-1,l=this.width/s,u=this.height/a;for(let p=0;p<t;p++){const m=p%this.verticesX,_=p/this.verticesX|0;r.push(m*l,_*u),n.push(m/s,_/a)}const c=s*a;for(let p=0;p<c;p++){const m=p%s,_=p/s|0,C=_*this.verticesX+m,b=_*this.verticesX+m+1,v=(_+1)*this.verticesX+m,w=(_+1)*this.verticesX+m+1;i.push(C,b,v,b,w,v)}this.buffers[0].data=new Float32Array(r),this.buffers[1].data=new Float32Array(n),this.indexBuffer.data=new Uint32Array(i),this.buffers[0].update(),this.buffers[1].update(),this.indexBuffer.update()}};qe.defaultOptions={width:100,height:100,verticesX:10,verticesY:10};let Yt=qe;class ue{constructor(){this.batcherName="default",this.packAsQuad=!1,this.indexOffset=0,this.attributeOffset=0,this.roundPixels=0,this._batcher=null,this._batch=null,this._textureMatrixUpdateId=-1,this._uvUpdateId=-1}get blendMode(){return this.renderable.groupBlendMode}get topology(){return this._topology||this.geometry.topology}set topology(e){this._topology=e}reset(){this.renderable=null,this.texture=null,this._batcher=null,this._batch=null,this.geometry=null,this._uvUpdateId=-1,this._textureMatrixUpdateId=-1}setTexture(e){this.texture!==e&&(this.texture=e,this._textureMatrixUpdateId=-1)}get uvs(){const t=this.geometry.getBuffer("aUV"),r=t.data;let n=r;const i=this.texture.textureMatrix;return i.isSimple||(n=this._transformedUvs,(this._textureMatrixUpdateId!==i._updateID||this._uvUpdateId!==t._updateID)&&((!n||n.length<r.length)&&(n=this._transformedUvs=new Float32Array(r.length)),this._textureMatrixUpdateId=i._updateID,this._uvUpdateId=t._updateID,i.multiplyUvs(r,n))),n}get positions(){return this.geometry.positions}get indices(){return this.geometry.indices}get color(){return this.renderable.groupColorAlpha}get groupTransform(){return this.renderable.groupTransform}get attributeSize(){return this.geometry.positions.length/2}get indexSize(){return this.geometry.indices.length}}class Se{destroy(){}}class Je{constructor(e,t){this.localUniforms=new z({uTransformMatrix:{value:new F,type:"mat3x3<f32>"},uColor:{value:new Float32Array([1,1,1,1]),type:"vec4<f32>"},uRound:{value:0,type:"f32"}}),this.localUniformsBindGroup=new Ae({0:this.localUniforms}),this.renderer=e,this._adaptor=t,this._adaptor.init()}validateRenderable(e){const t=this._getMeshData(e),r=t.batched,n=e.batched;if(t.batched=n,r!==n)return!0;if(n){const i=e._geometry;if(i.indices.length!==t.indexSize||i.positions.length!==t.vertexSize)return t.indexSize=i.indices.length,t.vertexSize=i.positions.length,!0;const s=this._getBatchableMesh(e);return s.texture.uid!==e._texture.uid&&(s._textureMatrixUpdateId=-1),!s._batcher.checkAndUpdateTexture(s,e._texture)}return!1}addRenderable(e,t){var i,s;const r=this.renderer.renderPipes.batch,n=this._getMeshData(e);if(e.didViewUpdate&&(n.indexSize=(i=e._geometry.indices)==null?void 0:i.length,n.vertexSize=(s=e._geometry.positions)==null?void 0:s.length),n.batched){const a=this._getBatchableMesh(e);a.setTexture(e._texture),a.geometry=e._geometry,r.addToBatch(a,t)}else r.break(t),t.add(e)}updateRenderable(e){if(e.batched){const t=this._getBatchableMesh(e);t.setTexture(e._texture),t.geometry=e._geometry,t._batcher.updateElement(t)}}execute(e){if(!e.isRenderable)return;e.state.blendMode=se(e.groupBlendMode,e.texture._source);const t=this.localUniforms;t.uniforms.uTransformMatrix=e.groupTransform,t.uniforms.uRound=this.renderer._roundPixels|e._roundPixels,t.update(),N(e.groupColorAlpha,t.uniforms.uColor,0),this._adaptor.execute(this,e)}_getMeshData(e){var t,r;return(t=e._gpuData)[r=this.renderer.uid]||(t[r]=new Se),e._gpuData[this.renderer.uid].meshData||this._initMeshData(e)}_initMeshData(e){return e._gpuData[this.renderer.uid].meshData={batched:e.batched,indexSize:0,vertexSize:0},e._gpuData[this.renderer.uid].meshData}_getBatchableMesh(e){var t,r;return(t=e._gpuData)[r=this.renderer.uid]||(t[r]=new Se),e._gpuData[this.renderer.uid].batchableMesh||this._initBatchableMesh(e)}_initBatchableMesh(e){const t=new ue;return t.renderable=e,t.setTexture(e._texture),t.transform=e.groupTransform,t.roundPixels=this.renderer._roundPixels|e._roundPixels,e._gpuData[this.renderer.uid].batchableMesh=t,t}destroy(){this.localUniforms=null,this.localUniformsBindGroup=null,this._adaptor.destroy(),this._adaptor=null,this.renderer=null}}Je.extension={type:[g.WebGLPipes,g.WebGPUPipes,g.CanvasPipes],name:"mesh"};class Xt{execute(e,t){const r=e.state,n=e.renderer,i=t.shader||e.defaultShader;i.resources.uTexture=t.texture._source,i.resources.uniforms=e.localUniforms;const s=n.gl,a=e.getBuffers(t);n.shader.bind(i),n.state.set(r),n.geometry.bind(a.geometry,i.glProgram);const u=a.geometry.indexBuffer.data.BYTES_PER_ELEMENT===2?s.UNSIGNED_SHORT:s.UNSIGNED_INT;s.drawElements(s.TRIANGLES,t.particleChildren.length*6,u,0)}}class Ht{execute(e,t){const r=e.renderer,n=t.shader||e.defaultShader;n.groups[0]=r.renderPipes.uniformBatch.getUniformBindGroup(e.localUniforms,!0),n.groups[1]=r.texture.getTextureBindGroup(t.texture);const i=e.state,s=e.getBuffers(t);r.encoder.draw({geometry:s.geometry,shader:t.shader||e.defaultShader,state:i,size:t.particleChildren.length*6})}}function Fe(o,e=null){const t=o*6;if(t>65535?e||(e=new Uint32Array(t)):e||(e=new Uint16Array(t)),e.length!==t)throw new Error(`Out buffer length is incorrect, got ${e.length} and expected ${t}`);for(let r=0,n=0;r<t;r+=6,n+=4)e[r+0]=n+0,e[r+1]=n+1,e[r+2]=n+2,e[r+3]=n+0,e[r+4]=n+2,e[r+5]=n+3;return e}function Kt(o){return{dynamicUpdate:Re(o,!0),staticUpdate:Re(o,!1)}}function Re(o,e){const t=[];t.push(`

        var index = 0;

        for (let i = 0; i < ps.length; ++i)
        {
            const p = ps[i];

            `);let r=0;for(const i in o){const s=o[i];if(e!==s.dynamic)continue;t.push(`offset = index + ${r}`),t.push(s.code);const a=Z(s.format);r+=a.stride/4}t.push(`
            index += stride * 4;
        }
    `),t.unshift(`
        var stride = ${r};
    `);const n=t.join(`
`);return new Function("ps","f32v","u32v",n)}class Nt{constructor(e){var c;this._size=0,this._generateParticleUpdateCache={};const t=this._size=(c=e.size)!=null?c:1e3,r=e.properties;let n=0,i=0;for(const h in r){const d=r[h],f=Z(d.format);d.dynamic?i+=f.stride:n+=f.stride}this._dynamicStride=i/4,this._staticStride=n/4,this.staticAttributeBuffer=new V(t*4*n),this.dynamicAttributeBuffer=new V(t*4*i),this.indexBuffer=Fe(t);const s=new ne;let a=0,l=0;this._staticBuffer=new E({data:new Float32Array(1),label:"static-particle-buffer",shrinkToFit:!1,usage:S.VERTEX|S.COPY_DST}),this._dynamicBuffer=new E({data:new Float32Array(1),label:"dynamic-particle-buffer",shrinkToFit:!1,usage:S.VERTEX|S.COPY_DST});for(const h in r){const d=r[h],f=Z(d.format);d.dynamic?(s.addAttribute(d.attributeName,{buffer:this._dynamicBuffer,stride:this._dynamicStride*4,offset:a*4,format:d.format}),a+=f.size):(s.addAttribute(d.attributeName,{buffer:this._staticBuffer,stride:this._staticStride*4,offset:l*4,format:d.format}),l+=f.size)}s.addIndex(this.indexBuffer);const u=this.getParticleUpdate(r);this._dynamicUpload=u.dynamicUpdate,this._staticUpload=u.staticUpdate,this.geometry=s}getParticleUpdate(e){const t=jt(e);return this._generateParticleUpdateCache[t]?this._generateParticleUpdateCache[t]:(this._generateParticleUpdateCache[t]=this.generateParticleUpdate(e),this._generateParticleUpdateCache[t])}generateParticleUpdate(e){return Kt(e)}update(e,t){e.length>this._size&&(t=!0,this._size=Math.max(e.length,this._size*1.5|0),this.staticAttributeBuffer=new V(this._size*this._staticStride*4*4),this.dynamicAttributeBuffer=new V(this._size*this._dynamicStride*4*4),this.indexBuffer=Fe(this._size),this.geometry.indexBuffer.setDataWithSize(this.indexBuffer,this.indexBuffer.byteLength,!0));const r=this.dynamicAttributeBuffer;if(this._dynamicUpload(e,r.float32View,r.uint32View),this._dynamicBuffer.setDataWithSize(this.dynamicAttributeBuffer.float32View,e.length*this._dynamicStride*4,!0),t){const n=this.staticAttributeBuffer;this._staticUpload(e,n.float32View,n.uint32View),this._staticBuffer.setDataWithSize(n.float32View,e.length*this._staticStride*4,!0)}}destroy(){this._staticBuffer.destroy(),this._dynamicBuffer.destroy(),this.geometry.destroy()}}function jt(o){const e=[];for(const t in o){const r=o[t];e.push(t,r.code,r.dynamic?"d":"s")}return e.join("_")}var $t=`varying vec2 vUV;
varying vec4 vColor;

uniform sampler2D uTexture;

void main(void){
    vec4 color = texture2D(uTexture, vUV) * vColor;
    gl_FragColor = color;
}`,qt=`attribute vec2 aVertex;
attribute vec2 aUV;
attribute vec4 aColor;

attribute vec2 aPosition;
attribute float aRotation;

uniform mat3 uTranslationMatrix;
uniform float uRound;
uniform vec2 uResolution;
uniform vec4 uColor;

varying vec2 vUV;
varying vec4 vColor;

vec2 roundPixels(vec2 position, vec2 targetSize)
{       
    return (floor(((position * 0.5 + 0.5) * targetSize) + 0.5) / targetSize) * 2.0 - 1.0;
}

void main(void){
    float cosRotation = cos(aRotation);
    float sinRotation = sin(aRotation);
    float x = aVertex.x * cosRotation - aVertex.y * sinRotation;
    float y = aVertex.x * sinRotation + aVertex.y * cosRotation;

    vec2 v = vec2(x, y);
    v = v + aPosition;

    gl_Position = vec4((uTranslationMatrix * vec3(v, 1.0)).xy, 0.0, 1.0);

    if(uRound == 1.0)
    {
        gl_Position.xy = roundPixels(gl_Position.xy, uResolution);
    }

    vUV = aUV;
    vColor = vec4(aColor.rgb * aColor.a, aColor.a) * uColor;
}
`,Ue=`
struct ParticleUniforms {
  uTranslationMatrix:mat3x3<f32>,
  uColor:vec4<f32>,
  uRound:f32,
  uResolution:vec2<f32>,
};

fn roundPixels(position: vec2<f32>, targetSize: vec2<f32>) -> vec2<f32>
{
  return (floor(((position * 0.5 + 0.5) * targetSize) + 0.5) / targetSize) * 2.0 - 1.0;
}

@group(0) @binding(0) var<uniform> uniforms: ParticleUniforms;

@group(1) @binding(0) var uTexture: texture_2d<f32>;
@group(1) @binding(1) var uSampler : sampler;

struct VSOutput {
    @builtin(position) position: vec4<f32>,
    @location(0) uv : vec2<f32>,
    @location(1) color : vec4<f32>,
  };
@vertex
fn mainVertex(
  @location(0) aVertex: vec2<f32>,
  @location(1) aPosition: vec2<f32>,
  @location(2) aUV: vec2<f32>,
  @location(3) aColor: vec4<f32>,
  @location(4) aRotation: f32,
) -> VSOutput {
  
   let v = vec2(
       aVertex.x * cos(aRotation) - aVertex.y * sin(aRotation),
       aVertex.x * sin(aRotation) + aVertex.y * cos(aRotation)
   ) + aPosition;

   var position = vec4((uniforms.uTranslationMatrix * vec3(v, 1.0)).xy, 0.0, 1.0);

   if(uniforms.uRound == 1.0) {
       position = vec4(roundPixels(position.xy, uniforms.uResolution), position.zw);
   }

    let vColor = vec4(aColor.rgb * aColor.a, aColor.a) * uniforms.uColor;

  return VSOutput(
   position,
   aUV,
   vColor,
  );
}

@fragment
fn mainFragment(
  @location(0) uv: vec2<f32>,
  @location(1) color: vec4<f32>,
  @builtin(position) position: vec4<f32>,
) -> @location(0) vec4<f32> {

    var sample = textureSample(uTexture, uSampler, uv) * color;
   
    return sample;
}`;class Qt extends ae{constructor(){const e=Ge.from({vertex:qt,fragment:$t}),t=Me.from({fragment:{source:Ue,entryPoint:"mainFragment"},vertex:{source:Ue,entryPoint:"mainVertex"}});super({glProgram:e,gpuProgram:t,resources:{uTexture:B.WHITE.source,uSampler:new ee({}),uniforms:{uTranslationMatrix:{value:new F,type:"mat3x3<f32>"},uColor:{value:new ke(16777215),type:"vec4<f32>"},uRound:{value:1,type:"f32"},uResolution:{value:[0,0],type:"vec2<f32>"}}}})}}class Ze{constructor(e,t){this.state=H.for2d(),this.localUniforms=new z({uTranslationMatrix:{value:new F,type:"mat3x3<f32>"},uColor:{value:new Float32Array(4),type:"vec4<f32>"},uRound:{value:1,type:"f32"},uResolution:{value:[0,0],type:"vec2<f32>"}}),this.renderer=e,this.adaptor=t,this.defaultShader=new Qt,this.state=H.for2d()}validateRenderable(e){return!1}addRenderable(e,t){this.renderer.renderPipes.batch.break(t),t.add(e)}getBuffers(e){return e._gpuData[this.renderer.uid]||this._initBuffer(e)}_initBuffer(e){return e._gpuData[this.renderer.uid]=new Nt({size:e.particleChildren.length,properties:e._properties}),e._gpuData[this.renderer.uid]}updateRenderable(e){}execute(e){const t=e.particleChildren;if(t.length===0)return;const r=this.renderer,n=this.getBuffers(e);e.texture||(e.texture=t[0].texture);const i=this.state;n.update(t,e._childrenDirty),e._childrenDirty=!1,i.blendMode=se(e.blendMode,e.texture._source);const s=this.localUniforms.uniforms,a=s.uTranslationMatrix;e.worldTransform.copyTo(a),a.prepend(r.globalUniforms.globalUniformData.projectionMatrix),s.uResolution=r.globalUniforms.globalUniformData.resolution,s.uRound=r._roundPixels|e._roundPixels,N(e.groupColorAlpha,s.uColor,0),this.adaptor.execute(this,e)}destroy(){this.renderer=null,this.defaultShader&&(this.defaultShader.destroy(),this.defaultShader=null)}}class et extends Ze{constructor(e){super(e,new Xt)}}et.extension={type:[g.WebGLPipes],name:"particle"};class tt extends Ze{constructor(e){super(e,new Ht)}}tt.extension={type:[g.WebGPUPipes],name:"particle"};const rt=class nt extends Yt{constructor(e={}){e={...nt.defaultOptions,...e},super({width:e.width,height:e.height,verticesX:4,verticesY:4}),this.update(e)}update(e){var t,r,n,i,s,a,l,u,c,h;this.width=(t=e.width)!=null?t:this.width,this.height=(r=e.height)!=null?r:this.height,this._originalWidth=(n=e.originalWidth)!=null?n:this._originalWidth,this._originalHeight=(i=e.originalHeight)!=null?i:this._originalHeight,this._leftWidth=(s=e.leftWidth)!=null?s:this._leftWidth,this._rightWidth=(a=e.rightWidth)!=null?a:this._rightWidth,this._topHeight=(l=e.topHeight)!=null?l:this._topHeight,this._bottomHeight=(u=e.bottomHeight)!=null?u:this._bottomHeight,this._anchorX=(c=e.anchor)==null?void 0:c.x,this._anchorY=(h=e.anchor)==null?void 0:h.y,this.updateUvs(),this.updatePositions()}updatePositions(){const e=this.positions,{width:t,height:r,_leftWidth:n,_rightWidth:i,_topHeight:s,_bottomHeight:a,_anchorX:l,_anchorY:u}=this,c=n+i,h=t>c?1:t/c,d=s+a,f=r>d?1:r/d,x=Math.min(h,f),p=l*t,m=u*r;e[0]=e[8]=e[16]=e[24]=-p,e[2]=e[10]=e[18]=e[26]=n*x-p,e[4]=e[12]=e[20]=e[28]=t-i*x-p,e[6]=e[14]=e[22]=e[30]=t-p,e[1]=e[3]=e[5]=e[7]=-m,e[9]=e[11]=e[13]=e[15]=s*x-m,e[17]=e[19]=e[21]=e[23]=r-a*x-m,e[25]=e[27]=e[29]=e[31]=r-m,this.getBuffer("aPosition").update()}updateUvs(){const e=this.uvs;e[0]=e[8]=e[16]=e[24]=0,e[1]=e[3]=e[5]=e[7]=0,e[6]=e[14]=e[22]=e[30]=1,e[25]=e[27]=e[29]=e[31]=1;const t=1/this._originalWidth,r=1/this._originalHeight;e[2]=e[10]=e[18]=e[26]=t*this._leftWidth,e[9]=e[11]=e[13]=e[15]=r*this._topHeight,e[4]=e[12]=e[20]=e[28]=1-t*this._rightWidth,e[17]=e[19]=e[21]=e[23]=1-r*this._bottomHeight,this.getBuffer("aUV").update()}};rt.defaultOptions={width:100,height:100,leftWidth:10,topHeight:10,rightWidth:10,bottomHeight:10,originalWidth:100,originalHeight:100};let Jt=rt;class Zt extends ue{constructor(){super(),this.geometry=new Jt}destroy(){this.geometry.destroy()}}class it{constructor(e){this._renderer=e}addRenderable(e,t){const r=this._getGpuSprite(e);e.didViewUpdate&&this._updateBatchableSprite(e,r),this._renderer.renderPipes.batch.addToBatch(r,t)}updateRenderable(e){const t=this._getGpuSprite(e);e.didViewUpdate&&this._updateBatchableSprite(e,t),t._batcher.updateElement(t)}validateRenderable(e){const t=this._getGpuSprite(e);return!t._batcher.checkAndUpdateTexture(t,e._texture)}_updateBatchableSprite(e,t){t.geometry.update(e),t.setTexture(e._texture)}_getGpuSprite(e){return e._gpuData[this._renderer.uid]||this._initGPUSprite(e)}_initGPUSprite(e){const t=e._gpuData[this._renderer.uid]=new Zt,r=t;return r.renderable=e,r.transform=e.groupTransform,r.texture=e._texture,r.roundPixels=this._renderer._roundPixels|e._roundPixels,e.didViewUpdate||this._updateBatchableSprite(e,r),t}destroy(){this._renderer=null}}it.extension={type:[g.WebGLPipes,g.WebGPUPipes,g.CanvasPipes],name:"nineSliceSprite"};const er={name:"tiling-bit",vertex:{header:`
            struct TilingUniforms {
                uMapCoord:mat3x3<f32>,
                uClampFrame:vec4<f32>,
                uClampOffset:vec2<f32>,
                uTextureTransform:mat3x3<f32>,
                uSizeAnchor:vec4<f32>
            };

            @group(2) @binding(0) var<uniform> tilingUniforms: TilingUniforms;
            @group(2) @binding(1) var uTexture: texture_2d<f32>;
            @group(2) @binding(2) var uSampler: sampler;
        `,main:`
            uv = (tilingUniforms.uTextureTransform * vec3(uv, 1.0)).xy;

            position = (position - tilingUniforms.uSizeAnchor.zw) * tilingUniforms.uSizeAnchor.xy;
        `},fragment:{header:`
            struct TilingUniforms {
                uMapCoord:mat3x3<f32>,
                uClampFrame:vec4<f32>,
                uClampOffset:vec2<f32>,
                uTextureTransform:mat3x3<f32>,
                uSizeAnchor:vec4<f32>
            };

            @group(2) @binding(0) var<uniform> tilingUniforms: TilingUniforms;
            @group(2) @binding(1) var uTexture: texture_2d<f32>;
            @group(2) @binding(2) var uSampler: sampler;
        `,main:`

            var coord = vUV + ceil(tilingUniforms.uClampOffset - vUV);
            coord = (tilingUniforms.uMapCoord * vec3(coord, 1.0)).xy;
            var unclamped = coord;
            coord = clamp(coord, tilingUniforms.uClampFrame.xy, tilingUniforms.uClampFrame.zw);

            var bias = 0.;

            if(unclamped.x == coord.x && unclamped.y == coord.y)
            {
                bias = -32.;
            }

            outColor = textureSampleBias(uTexture, uSampler, coord, bias);
        `}},tr={name:"tiling-bit",vertex:{header:`
            uniform mat3 uTextureTransform;
            uniform vec4 uSizeAnchor;

        `,main:`
            uv = (uTextureTransform * vec3(aUV, 1.0)).xy;

            position = (position - uSizeAnchor.zw) * uSizeAnchor.xy;
        `},fragment:{header:`
            uniform sampler2D uTexture;
            uniform mat3 uMapCoord;
            uniform vec4 uClampFrame;
            uniform vec2 uClampOffset;
        `,main:`

        vec2 coord = vUV + ceil(uClampOffset - vUV);
        coord = (uMapCoord * vec3(coord, 1.0)).xy;
        vec2 unclamped = coord;
        coord = clamp(coord, uClampFrame.xy, uClampFrame.zw);

        outColor = texture(uTexture, coord, unclamped == coord ? 0.0 : -32.0);// lod-bias very negative to force lod 0

        `}};let D,k;class rr extends ae{constructor(){D!=null||(D=Oe({name:"tiling-sprite-shader",bits:[Bt,er,Ie]})),k!=null||(k=Ee({name:"tiling-sprite-shader",bits:[Mt,tr,Ve]}));const e=new z({uMapCoord:{value:new F,type:"mat3x3<f32>"},uClampFrame:{value:new Float32Array([0,0,1,1]),type:"vec4<f32>"},uClampOffset:{value:new Float32Array([0,0]),type:"vec2<f32>"},uTextureTransform:{value:new F,type:"mat3x3<f32>"},uSizeAnchor:{value:new Float32Array([100,100,.5,.5]),type:"vec4<f32>"}});super({glProgram:k,gpuProgram:D,resources:{localUniforms:new z({uTransformMatrix:{value:new F,type:"mat3x3<f32>"},uColor:{value:new Float32Array([1,1,1,1]),type:"vec4<f32>"},uRound:{value:0,type:"f32"}}),tilingUniforms:e,uTexture:B.EMPTY.source,uSampler:B.EMPTY.source.style}})}updateUniforms(e,t,r,n,i,s){const a=this.resources.tilingUniforms,l=s.width,u=s.height,c=s.textureMatrix,h=a.uniforms.uTextureTransform;h.set(r.a*l/e,r.b*l/t,r.c*u/e,r.d*u/t,r.tx/e,r.ty/t),h.invert(),a.uniforms.uMapCoord=c.mapCoord,a.uniforms.uClampFrame=c.uClampFrame,a.uniforms.uClampOffset=c.uClampOffset,a.uniforms.uTextureTransform=h,a.uniforms.uSizeAnchor[0]=e,a.uniforms.uSizeAnchor[1]=t,a.uniforms.uSizeAnchor[2]=n,a.uniforms.uSizeAnchor[3]=i,s&&(this.resources.uTexture=s.source,this.resources.uSampler=s.source.style)}}class nr extends oe{constructor(){super({positions:new Float32Array([0,0,1,0,1,1,0,1]),uvs:new Float32Array([0,0,1,0,1,1,0,1]),indices:new Uint32Array([0,1,2,0,2,3])})}}function ir(o,e){const t=o.anchor.x,r=o.anchor.y;e[0]=-t*o.width,e[1]=-r*o.height,e[2]=(1-t)*o.width,e[3]=-r*o.height,e[4]=(1-t)*o.width,e[5]=(1-r)*o.height,e[6]=-t*o.width,e[7]=(1-r)*o.height}function sr(o,e,t,r){let n=0;const i=o.length/(e||2),s=r.a,a=r.b,l=r.c,u=r.d,c=r.tx,h=r.ty;for(t*=e;n<i;){const d=o[t],f=o[t+1];o[t]=s*d+l*f+c,o[t+1]=a*d+u*f+h,t+=e,n++}}function ar(o,e){const t=o.texture,r=t.frame.width,n=t.frame.height;let i=0,s=0;o.applyAnchorToTexture&&(i=o.anchor.x,s=o.anchor.y),e[0]=e[6]=-i,e[2]=e[4]=1-i,e[1]=e[3]=-s,e[5]=e[7]=1-s;const a=F.shared;a.copyFrom(o._tileTransform.matrix),a.tx/=o.width,a.ty/=o.height,a.invert(),a.scale(o.width/r,o.height/n),sr(e,2,0,a)}const W=new nr;class or{constructor(){this.canBatch=!0,this.geometry=new oe({indices:W.indices.slice(),positions:W.positions.slice(),uvs:W.uvs.slice()})}destroy(){var e;this.geometry.destroy(),(e=this.shader)==null||e.destroy()}}class st{constructor(e){this._state=H.default2d,this._renderer=e}validateRenderable(e){const t=this._getTilingSpriteData(e),r=t.canBatch;this._updateCanBatch(e);const n=t.canBatch;if(n&&n===r){const{batchableMesh:i}=t;return!i._batcher.checkAndUpdateTexture(i,e.texture)}return r!==n}addRenderable(e,t){const r=this._renderer.renderPipes.batch;this._updateCanBatch(e);const n=this._getTilingSpriteData(e),{geometry:i,canBatch:s}=n;if(s){n.batchableMesh||(n.batchableMesh=new ue);const a=n.batchableMesh;e.didViewUpdate&&(this._updateBatchableMesh(e),a.geometry=i,a.renderable=e,a.transform=e.groupTransform,a.setTexture(e._texture)),a.roundPixels=this._renderer._roundPixels|e._roundPixels,r.addToBatch(a,t)}else r.break(t),n.shader||(n.shader=new rr),this.updateRenderable(e),t.add(e)}execute(e){const{shader:t}=this._getTilingSpriteData(e);t.groups[0]=this._renderer.globalUniforms.bindGroup;const r=t.resources.localUniforms.uniforms;r.uTransformMatrix=e.groupTransform,r.uRound=this._renderer._roundPixels|e._roundPixels,N(e.groupColorAlpha,r.uColor,0),this._state.blendMode=se(e.groupBlendMode,e.texture._source),this._renderer.encoder.draw({geometry:W,shader:t,state:this._state})}updateRenderable(e){const t=this._getTilingSpriteData(e),{canBatch:r}=t;if(r){const{batchableMesh:n}=t;e.didViewUpdate&&this._updateBatchableMesh(e),n._batcher.updateElement(n)}else if(e.didViewUpdate){const{shader:n}=t;n.updateUniforms(e.width,e.height,e._tileTransform.matrix,e.anchor.x,e.anchor.y,e.texture)}}_getTilingSpriteData(e){return e._gpuData[this._renderer.uid]||this._initTilingSpriteData(e)}_initTilingSpriteData(e){const t=new or;return t.renderable=e,e._gpuData[this._renderer.uid]=t,t}_updateBatchableMesh(e){const t=this._getTilingSpriteData(e),{geometry:r}=t,n=e.texture.source.style;n.addressMode!=="repeat"&&(n.addressMode="repeat",n.update()),ar(e,r.uvs),ir(e,r.positions)}destroy(){this._renderer=null}_updateCanBatch(e){const t=this._getTilingSpriteData(e),r=e.texture;let n=!0;return this._renderer.type===ie.WEBGL&&(n=this._renderer.context.supports.nonPowOf2wrapping),t.canBatch=r.textureMatrix.isSimple&&(n||r.source.isPowerOfTwo),t.canBatch}}st.extension={type:[g.WebGLPipes,g.WebGPUPipes,g.CanvasPipes],name:"tilingSprite"};const ur={name:"local-uniform-msdf-bit",vertex:{header:`
            struct LocalUniforms {
                uColor:vec4<f32>,
                uTransformMatrix:mat3x3<f32>,
                uDistance: f32,
                uRound:f32,
            }

            @group(2) @binding(0) var<uniform> localUniforms : LocalUniforms;
        `,main:`
            vColor *= localUniforms.uColor;
            modelMatrix *= localUniforms.uTransformMatrix;
        `,end:`
            if(localUniforms.uRound == 1)
            {
                vPosition = vec4(roundPixels(vPosition.xy, globalUniforms.uResolution), vPosition.zw);
            }
        `},fragment:{header:`
            struct LocalUniforms {
                uColor:vec4<f32>,
                uTransformMatrix:mat3x3<f32>,
                uDistance: f32
            }

            @group(2) @binding(0) var<uniform> localUniforms : LocalUniforms;
         `,main:`
            outColor = vec4<f32>(calculateMSDFAlpha(outColor, localUniforms.uColor, localUniforms.uDistance));
        `}},lr={name:"local-uniform-msdf-bit",vertex:{header:`
            uniform mat3 uTransformMatrix;
            uniform vec4 uColor;
            uniform float uRound;
        `,main:`
            vColor *= uColor;
            modelMatrix *= uTransformMatrix;
        `,end:`
            if(uRound == 1.)
            {
                gl_Position.xy = roundPixels(gl_Position.xy, uResolution);
            }
        `},fragment:{header:`
            uniform float uDistance;
         `,main:`
            outColor = vec4(calculateMSDFAlpha(outColor, vColor, uDistance));
        `}},cr={name:"msdf-bit",fragment:{header:`
            fn calculateMSDFAlpha(msdfColor:vec4<f32>, shapeColor:vec4<f32>, distance:f32) -> f32 {

                // MSDF
                var median = msdfColor.r + msdfColor.g + msdfColor.b -
                    min(msdfColor.r, min(msdfColor.g, msdfColor.b)) -
                    max(msdfColor.r, max(msdfColor.g, msdfColor.b));

                // SDF
                median = min(median, msdfColor.a);

                var screenPxDistance = distance * (median - 0.5);
                var alpha = clamp(screenPxDistance + 0.5, 0.0, 1.0);
                if (median < 0.01) {
                    alpha = 0.0;
                } else if (median > 0.99) {
                    alpha = 1.0;
                }

                // Gamma correction for coverage-like alpha
                var luma: f32 = dot(shapeColor.rgb, vec3<f32>(0.299, 0.587, 0.114));
                var gamma: f32 = mix(1.0, 1.0 / 2.2, luma);
                var coverage: f32 = pow(shapeColor.a * alpha, gamma);

                return coverage;

            }
        `}},dr={name:"msdf-bit",fragment:{header:`
            float calculateMSDFAlpha(vec4 msdfColor, vec4 shapeColor, float distance) {

                // MSDF
                float median = msdfColor.r + msdfColor.g + msdfColor.b -
                                min(msdfColor.r, min(msdfColor.g, msdfColor.b)) -
                                max(msdfColor.r, max(msdfColor.g, msdfColor.b));

                // SDF
                median = min(median, msdfColor.a);

                float screenPxDistance = distance * (median - 0.5);
                float alpha = clamp(screenPxDistance + 0.5, 0.0, 1.0);

                if (median < 0.01) {
                    alpha = 0.0;
                } else if (median > 0.99) {
                    alpha = 1.0;
                }

                // Gamma correction for coverage-like alpha
                float luma = dot(shapeColor.rgb, vec3(0.299, 0.587, 0.114));
                float gamma = mix(1.0, 1.0 / 2.2, luma);
                float coverage = pow(shapeColor.a * alpha, gamma);

                return coverage;
            }
        `}};let O,I;class hr extends ae{constructor(e){const t=new z({uColor:{value:new Float32Array([1,1,1,1]),type:"vec4<f32>"},uTransformMatrix:{value:new F,type:"mat3x3<f32>"},uDistance:{value:4,type:"f32"},uRound:{value:0,type:"f32"}});O!=null||(O=Oe({name:"sdf-shader",bits:[vt,bt(e),ur,cr,Ie]})),I!=null||(I=Ee({name:"sdf-shader",bits:[yt,Tt(e),lr,dr,Ve]})),super({glProgram:I,gpuProgram:O,resources:{localUniforms:t,batchSamplers:wt(e)}})}}class fr extends Ct{destroy(){this.context.customShader&&this.context.customShader.destroy(),super.destroy()}}class at{constructor(e){this._renderer=e}validateRenderable(e){const t=this._getGpuBitmapText(e);return this._renderer.renderPipes.graphics.validateRenderable(t)}addRenderable(e,t){const r=this._getGpuBitmapText(e);Be(e,r),e._didTextUpdate&&(e._didTextUpdate=!1,this._updateContext(e,r)),this._renderer.renderPipes.graphics.addRenderable(r,t),r.context.customShader&&this._updateDistanceField(e)}updateRenderable(e){const t=this._getGpuBitmapText(e);Be(e,t),this._renderer.renderPipes.graphics.updateRenderable(t),t.context.customShader&&this._updateDistanceField(e)}_updateContext(e,t){const{context:r}=t,n=Pt.getFont(e.text,e._style);r.clear(),n.distanceField.type!=="none"&&(r.customShader||(r.customShader=new hr(this._renderer.limits.maxBatchableTextures)));const i=A.graphemeSegmenter(e.text),s=e._style;let a=n.baseLineOffset;const l=St(i,s,n,!0),u=s.padding,c=l.scale;let h=l.width,d=l.height+l.offsetY;s._stroke&&(h+=s._stroke.width/c,d+=s._stroke.width/c),r.translate(-e._anchor._x*h-u,-e._anchor._y*d-u).scale(c,c);const f=n.applyFillAsTint?s._fill.color:16777215;let x=n.fontMetrics.fontSize,p=n.lineHeight;s.lineHeight&&(x=s.fontSize/c,p=s.lineHeight/c);let m=(p-x)/2;m-n.baseLineOffset<0&&(m=0);for(let _=0;_<l.lines.length;_++){const C=l.lines[_];for(let b=0;b<C.charPositions.length;b++){const v=C.chars[b],w=n.chars[v];if(w!=null&&w.texture){const G=w.texture;r.texture(G,f||"black",Math.round(C.charPositions[b]+w.xOffset),Math.round(a+w.yOffset+m),G.orig.width,G.orig.height)}}a+=p}}_getGpuBitmapText(e){return e._gpuData[this._renderer.uid]||this.initGpuText(e)}initGpuText(e){const t=new fr;return e._gpuData[this._renderer.uid]=t,this._updateContext(e,t),t}_updateDistanceField(e){const t=this._getGpuBitmapText(e).context,r=e._style.fontFamily,n=te.get(`${r}-bitmap`),{a:i,b:s,c:a,d:l}=e.groupTransform,u=Math.sqrt(i*i+s*s),c=Math.sqrt(a*a+l*l),h=(Math.abs(u)+Math.abs(c))/2,d=n.baseRenderedFontSize/e._style.fontSize,f=h*n.distanceField.range*(1/d);t.customShader.resources.localUniforms.uniforms.uDistance=f}destroy(){this._renderer=null}}at.extension={type:[g.WebGLPipes,g.WebGPUPipes,g.CanvasPipes],name:"bitmapText"};function Be(o,e){e.groupTransform=o.groupTransform,e.groupColorAlpha=o.groupColorAlpha,e.groupColor=o.groupColor,e.groupBlendMode=o.groupBlendMode,e.globalDisplayStatus=o.globalDisplayStatus,e.groupTransform=o.groupTransform,e.localDisplayStatus=o.localDisplayStatus,e.groupAlpha=o.groupAlpha,e._roundPixels=o._roundPixels}class pr extends We{constructor(e){super(),this.generatingTexture=!1,this.currentKey="--",this._renderer=e,e.runners.resolutionChange.add(this)}resolutionChange(){const e=this.renderable;e._autoResolution&&e.onViewUpdate()}destroy(){const{htmlText:e}=this._renderer;e.getReferenceCount(this.currentKey)===null?e.returnTexturePromise(this.texturePromise):e.decreaseReferenceCount(this.currentKey),this._renderer.runners.resolutionChange.remove(this),this.texturePromise=null,this._renderer=null}}function re(o,e){const{texture:t,bounds:r}=o,n=e._style._getFinalPadding();Ft(r,e._anchor,t);const i=e._anchor._x*n*2,s=e._anchor._y*n*2;r.minX-=n-i,r.minY-=n-s,r.maxX-=n-i,r.maxY-=n-s}class ot{constructor(e){this._renderer=e}validateRenderable(e){const t=this._getGpuText(e),r=e.styleKey;return t.currentKey!==r}addRenderable(e,t){const r=this._getGpuText(e);if(e._didTextUpdate){const n=e._autoResolution?this._renderer.resolution:e.resolution;(r.currentKey!==e.styleKey||e.resolution!==n)&&this._updateGpuText(e).catch(i=>{console.error(i)}),e._didTextUpdate=!1,re(r,e)}this._renderer.renderPipes.batch.addToBatch(r,t)}updateRenderable(e){const t=this._getGpuText(e);t._batcher.updateElement(t)}async _updateGpuText(e){e._didTextUpdate=!1;const t=this._getGpuText(e);if(t.generatingTexture)return;const r=t.texturePromise;t.texturePromise=null,t.generatingTexture=!0,e._resolution=e._autoResolution?this._renderer.resolution:e.resolution;let n=this._renderer.htmlText.getTexturePromise(e);r&&(n=n.finally(()=>{this._renderer.htmlText.decreaseReferenceCount(t.currentKey),this._renderer.htmlText.returnTexturePromise(r)})),t.texturePromise=n,t.currentKey=e.styleKey,t.texture=await n;const i=e.renderGroup||e.parentRenderGroup;i&&(i.structureDidChange=!0),t.generatingTexture=!1,re(t,e)}_getGpuText(e){return e._gpuData[this._renderer.uid]||this.initGpuText(e)}initGpuText(e){const t=new pr(this._renderer);return t.renderable=e,t.transform=e.groupTransform,t.texture=B.EMPTY,t.bounds={minX:0,maxX:1,minY:0,maxY:0},t.roundPixels=this._renderer._roundPixels|e._roundPixels,e._resolution=e._autoResolution?this._renderer.resolution:e.resolution,e._gpuData[this._renderer.uid]=t,t}destroy(){this._renderer=null}}ot.extension={type:[g.WebGLPipes,g.WebGPUPipes,g.CanvasPipes],name:"htmlText"};function mr(){const{userAgent:o}=K.get().getNavigator();return/^((?!chrome|android).)*safari/i.test(o)}const gr=new ze;function ut(o,e,t,r){const n=gr;n.minX=0,n.minY=0,n.maxX=o.width/r|0,n.maxY=o.height/r|0;const i=P.getOptimalTexture(n.width,n.height,r,!1);return i.source.uploadMethodId="image",i.source.resource=o,i.source.alphaMode="premultiply-alpha-on-upload",i.frame.width=e/r,i.frame.height=t/r,i.source.emit("update",i.source),i.updateUvs(),i}function xr(o,e){const t=e.fontFamily,r=[],n={},i=/font-family:([^;"\s]+)/g,s=o.match(i);function a(l){n[l]||(r.push(l),n[l]=!0)}if(Array.isArray(t))for(let l=0;l<t.length;l++)a(t[l]);else a(t);s&&s.forEach(l=>{const u=l.split(":")[1].trim();a(u)});for(const l in e.tagStyles){const u=e.tagStyles[l].fontFamily;a(u)}return r}async function _r(o){const t=await(await K.get().fetch(o)).blob(),r=new FileReader;return await new Promise((i,s)=>{r.onloadend=()=>i(r.result),r.onerror=s,r.readAsDataURL(t)})}async function vr(o,e){const t=await _r(e);return`@font-face {
        font-family: "${o.fontFamily}";
        font-weight: ${o.fontWeight};
        font-style: ${o.fontStyle};
        src: url('${t}');
    }`}const q=new Map;async function br(o){const e=o.filter(t=>te.has(`${t}-and-url`)).map(t=>{if(!q.has(t)){const{entries:r}=te.get(`${t}-and-url`),n=[];r.forEach(i=>{const s=i.url,l=i.faces.map(u=>({weight:u.weight,style:u.style}));n.push(...l.map(u=>vr({fontWeight:u.weight,fontStyle:u.style,fontFamily:t},s)))}),q.set(t,Promise.all(n).then(i=>i.join(`
`)))}return q.get(t)});return(await Promise.all(e)).join(`
`)}function yr(o,e,t,r,n){const{domElement:i,styleElement:s,svgRoot:a}=n;i.innerHTML=`<style>${e.cssStyle}</style><div style='padding:0;'>${o}</div>`,i.setAttribute("style",`transform: scale(${t});transform-origin: top left; display: inline-block`),s.textContent=r;const{width:l,height:u}=n.image;return a.setAttribute("width",l.toString()),a.setAttribute("height",u.toString()),new XMLSerializer().serializeToString(a)}function Tr(o,e){const t=Y.getOptimalCanvasAndContext(o.width,o.height,e),{context:r}=t;return r.clearRect(0,0,o.width,o.height),r.drawImage(o,0,0),t}function wr(o,e,t){return new Promise(async r=>{t&&await new Promise(n=>setTimeout(n,100)),o.onload=()=>{r()},o.src=`data:image/svg+xml;charset=utf8,${encodeURIComponent(e)}`,o.crossOrigin="anonymous"})}class lt{constructor(e){this._activeTextures={},this._renderer=e,this._createCanvas=e.type===ie.WEBGPU}getTexture(e){return this.getTexturePromise(e)}getManagedTexture(e){const t=e.styleKey;if(this._activeTextures[t])return this._increaseReferenceCount(t),this._activeTextures[t].promise;const r=this._buildTexturePromise(e).then(n=>(this._activeTextures[t].texture=n,n));return this._activeTextures[t]={texture:null,promise:r,usageCount:1},r}getReferenceCount(e){var t,r;return(r=(t=this._activeTextures[e])==null?void 0:t.usageCount)!=null?r:null}_increaseReferenceCount(e){this._activeTextures[e].usageCount++}decreaseReferenceCount(e){const t=this._activeTextures[e];!t||(t.usageCount--,t.usageCount===0&&(t.texture?this._cleanUp(t.texture):t.promise.then(r=>{t.texture=r,this._cleanUp(t.texture)}).catch(()=>{Q("HTMLTextSystem: Failed to clean texture")}),this._activeTextures[e]=null))}getTexturePromise(e){return this._buildTexturePromise(e)}async _buildTexturePromise(e){const{text:t,style:r,resolution:n,textureStyle:i}=e,s=X.get(je),a=xr(t,r),l=await br(a),u=Wt(t,r,l,s),c=Math.ceil(Math.ceil(Math.max(1,u.width)+r.padding*2)*n),h=Math.ceil(Math.ceil(Math.max(1,u.height)+r.padding*2)*n),d=s.image,f=2;d.width=(c|0)+f,d.height=(h|0)+f;const x=yr(t,r,n,l,s);await wr(d,x,mr()&&a.length>0);const p=d;let m;this._createCanvas&&(m=Tr(d,n));const _=ut(m?m.canvas:p,d.width-f,d.height-f,n);return i&&(_.source.style=i),this._createCanvas&&(this._renderer.texture.initSource(_.source),Y.returnCanvasAndContext(m)),X.return(s),_}returnTexturePromise(e){e.then(t=>{this._cleanUp(t)}).catch(()=>{Q("HTMLTextSystem: Failed to clean texture")})}_cleanUp(e){P.returnTexture(e,!0),e.source.resource=null,e.source.uploadMethodId="unknown"}destroy(){this._renderer=null;for(const e in this._activeTextures)this._activeTextures[e]&&this.returnTexturePromise(this._activeTextures[e].promise);this._activeTextures=null}}lt.extension={type:[g.WebGLSystem,g.WebGPUSystem,g.CanvasSystem],name:"htmlText"};class Cr extends We{constructor(e){super(),this._renderer=e,e.runners.resolutionChange.add(this)}resolutionChange(){const e=this.renderable;e._autoResolution&&e.onViewUpdate()}destroy(){const{canvasText:e}=this._renderer;e.getReferenceCount(this.currentKey)>0?e.decreaseReferenceCount(this.currentKey):this.texture&&e.returnTexture(this.texture),this._renderer.runners.resolutionChange.remove(this),this._renderer=null}}class ct{constructor(e){this._renderer=e}validateRenderable(e){const t=this._getGpuText(e),r=e.styleKey;return t.currentKey!==r?!0:e._didTextUpdate}addRenderable(e,t){const r=this._getGpuText(e);if(e._didTextUpdate){const n=e._autoResolution?this._renderer.resolution:e.resolution;(r.currentKey!==e.styleKey||e.resolution!==n)&&this._updateGpuText(e),e._didTextUpdate=!1,re(r,e)}this._renderer.renderPipes.batch.addToBatch(r,t)}updateRenderable(e){const t=this._getGpuText(e);t._batcher.updateElement(t)}_updateGpuText(e){const t=this._getGpuText(e);t.texture&&this._renderer.canvasText.decreaseReferenceCount(t.currentKey),e._resolution=e._autoResolution?this._renderer.resolution:e.resolution,t.texture=this._renderer.canvasText.getManagedTexture(e),t.currentKey=e.styleKey}_getGpuText(e){return e._gpuData[this._renderer.uid]||this.initGpuText(e)}initGpuText(e){const t=new Cr(this._renderer);return t.currentKey="--",t.renderable=e,t.transform=e.groupTransform,t.bounds={minX:0,maxX:1,minY:0,maxY:0},t.roundPixels=this._renderer._roundPixels|e._roundPixels,e._gpuData[this._renderer.uid]=t,t}destroy(){this._renderer=null}}ct.extension={type:[g.WebGLPipes,g.WebGPUPipes,g.CanvasPipes],name:"text"};class dt{constructor(e){this._activeTextures={},this._renderer=e}getTexture(e,t,r,n){var d;typeof e=="string"&&(L("8.0.0","CanvasTextSystem.getTexture: Use object TextOptions instead of separate arguments"),e={text:e,style:r,resolution:t}),e.style instanceof xe||(e.style=new xe(e.style)),e.textureStyle instanceof ee||(e.textureStyle=new ee(e.textureStyle)),typeof e.text!="string"&&(e.text=e.text.toString());const{text:i,style:s,textureStyle:a}=e,l=(d=e.resolution)!=null?d:this._renderer.resolution,{frame:u,canvasAndContext:c}=$.getCanvasAndContext({text:i,style:s,resolution:l}),h=ut(c.canvas,u.width,u.height,l);if(a&&(h.source.style=a),s.trim&&(u.pad(s.padding),h.frame.copyFrom(u),h.frame.scale(1/l),h.updateUvs()),s.filters){const f=this._applyFilters(h,s.filters);return this.returnTexture(h),$.returnCanvasAndContext(c),f}return this._renderer.texture.initSource(h._source),$.returnCanvasAndContext(c),h}returnTexture(e){const t=e.source;t.resource=null,t.uploadMethodId="unknown",t.alphaMode="no-premultiply-alpha",P.returnTexture(e,!0)}renderTextToCanvas(){L("8.10.0","CanvasTextSystem.renderTextToCanvas: no longer supported, use CanvasTextSystem.getTexture instead")}getManagedTexture(e){e._resolution=e._autoResolution?this._renderer.resolution:e.resolution;const t=e.styleKey;if(this._activeTextures[t])return this._increaseReferenceCount(t),this._activeTextures[t].texture;const r=this.getTexture({text:e.text,style:e.style,resolution:e._resolution,textureStyle:e.textureStyle});return this._activeTextures[t]={texture:r,usageCount:1},r}decreaseReferenceCount(e){const t=this._activeTextures[e];t.usageCount--,t.usageCount===0&&(this.returnTexture(t.texture),this._activeTextures[e]=null)}getReferenceCount(e){var t,r;return(r=(t=this._activeTextures[e])==null?void 0:t.usageCount)!=null?r:0}_increaseReferenceCount(e){this._activeTextures[e].usageCount++}_applyFilters(e,t){const r=this._renderer.renderTarget.renderTarget,n=this._renderer.filter.generateFilteredTexture({texture:e,filters:t});return this._renderer.renderTarget.bind(r,!1),n}destroy(){this._renderer=null;for(const e in this._activeTextures)this._activeTextures[e]&&this.returnTexture(this._activeTextures[e].texture);this._activeTextures=null}}dt.extension={type:[g.WebGLSystem,g.WebGPUSystem,g.CanvasSystem],name:"canvasText"};T.add(Le);T.add(Ye);T.add($e);T.add(Rt);T.add(Je);T.add(et);T.add(tt);T.add(dt);T.add(ct);T.add(at);T.add(lt);T.add(ot);T.add(st);T.add(it);T.add(He);T.add(Xe);
