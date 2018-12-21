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
    nunjucks = require('gulp-nunjucks');

abs = path.join(process.cwd(), 'assets');

// Clean up artifacts
const clean = function() {
    return del(['./dist/**/*'])
};

// Minify images and output to ./dist folder
const img = function() {
    return gulp.src('./assets/**/*.{png,ico}')
        .pipe(imagemin())
        .pipe(size())
        .pipe(gulp.dest('./dist/assets/'))
};

// Minify scripts, build manifest.json and output to ./dist folder
const scripts = function() {
    return gulp.src(['./assets/js/**/*.js', './assets/css/**/*.css'], {base: abs})
        .pipe(gulpif(/js$/, uglify()))
        .pipe(gulpif(/css$/, autoprefixer()))
        .pipe(gulpif(/css$/, cleancss()))
        .pipe(rev())
        .pipe(size())
        .pipe(gulp.dest('./dist/assets'))
        .pipe(rev.manifest('manifest.json', {merge: true}))
        .pipe(gulp.dest('./dist/assets'));
};

// Rewrite occurrences of scripts in template files
const patch = function() {
    var manifest = gulp.src('./dist/assets/manifest.json');
    return gulp.src('./templates/[^_]*.html')
        .pipe(nunjucks.compile())
        .pipe(revreplace({manifest: manifest}))
        .pipe(gulp.dest('./dist/html/'))
};

exports.default = gulp.series(
    clean,
    gulp.parallel(
        img,
        gulp.series(scripts, patch)
    ),
);