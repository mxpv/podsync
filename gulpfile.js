const gulp = require('gulp'),
    del = require('del'),
    path = require('path'),
    uglify = require('gulp-uglify'),
    rev = require('gulp-rev'),
    revreplace = require('gulp-rev-replace'),
    cleancss = require('gulp-clean-css'),
    autoprefixer = require('gulp-autoprefixer'),
    size = require('gulp-size'),
    gulpif = require('gulp-if'),
    imagemin = require('gulp-imagemin'),
    nunjucks = require('gulp-nunjucks'),
    browserSync = require('browser-sync').create();

abs = path.join(process.cwd(), 'assets');

const config = {
    images: './assets/**/*.{png,ico}',
    scripts: ['./assets/js/**/*.js', './assets/css/**/*.css'],
    html: './templates/[^_]*.html',

    dist: {
        assets: './dist/assets/',
        html: './dist/html/',
    },
};

// Clean up artifacts
const clean = function() {
    return del(['./dist/**/*'])
};

// Minify images and output to ./dist folder
const img = function() {
    return gulp.src(config.images)
        .pipe(imagemin())
        .pipe(size())
        .pipe(gulp.dest(config.dist.assets))
};

// Minify scripts, build manifest.json and output to ./dist folder
const scripts = function() {
    return gulp.src(config.scripts, {base: abs})
        .pipe(gulpif(/js$/, uglify()))
        .pipe(gulpif(/css$/, autoprefixer()))
        .pipe(gulpif(/css$/, cleancss()))
        .pipe(rev())
        .pipe(size())
        .pipe(gulp.dest(config.dist.assets))
        .pipe(rev.manifest('manifest.json', {merge: true}))
        .pipe(gulp.dest(config.dist.assets));
};

// Rewrite occurrences of scripts in template files
const html = function() {
    const manifest = gulp.src('./dist/assets/manifest.json');
    return gulp.src(config.html)
        .pipe(nunjucks.compile())
        .pipe(revreplace({manifest: manifest}))
        .pipe(gulp.dest(config.dist.html))
};

const dev = function(cb) {
    browserSync.init({
        server: ['./dist/html/', './dist/'],
        port: 8080,
        watch: true,
    });

    gulp.watch(config.images, img);
    gulp.watch(config.scripts, gulp.series(scripts, html));
    gulp.watch(config.html, html);

    cb();
};

const build = gulp.series(
    clean,
    gulp.parallel(
        img,
        gulp.series(scripts, html)
    ),
);

exports.default = build;
exports.dev = gulp.series(build, dev);