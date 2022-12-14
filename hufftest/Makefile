# TODO: This needs to build the kotlin version somehow
# TODO: This needs to run the ada version
# TODO: These all need to run on shakespeare.txt, it exposes many more bugs (multiple blocks, full blocks, ...)

.PHONY: all compress huff-rs huff-kt huff-go huff-ada

all: compress

compress: huff-rs huff-kt huff-go
	../huff-go/huffer compress -o go.huff test.txt
	../huff-rs/target/debug/huffrs compress -o rust.huff test.txt
	/Library/Java/JavaVirtualMachines/liberica-jdk-11.jdk/Contents/Home/bin/java "-javaagent:/Applications/IntelliJ IDEA CE.app/Contents/lib/idea_rt.jar=50387:/Applications/IntelliJ IDEA CE.app/Contents/bin" -Dfile.encoding=UTF-8 -classpath "/Users/lth/p/sandbox/huff-kt/out/production/huff-kt:/Users/lth/.m2/repository/org/jetbrains/kotlin/kotlin-stdlib-jdk8/1.7.20/kotlin-stdlib-jdk8-1.7.20.jar:/Users/lth/.m2/repository/org/jetbrains/kotlin/kotlin-stdlib/1.7.20/kotlin-stdlib-1.7.20.jar:/Users/lth/.m2/repository/org/jetbrains/kotlin/kotlin-stdlib-common/1.7.20/kotlin-stdlib-common-1.7.20.jar:/Users/lth/.m2/repository/org/jetbrains/annotations/13.0/annotations-13.0.jar:/Users/lth/.m2/repository/org/jetbrains/kotlin/kotlin-stdlib-jdk7/1.7.20/kotlin-stdlib-jdk7-1.7.20.jar" MainKt compress -o kotlin.huff test.txt
	cmp go.huff rust.huff
	cmp go.huff kotlin.huff

huff-rs:
	( cd ../huff-rs ; cargo build )

huff-go:
	( cd ../huff-go ; go build )

huff-kt:
	( cd ../huff-kt ; test out/production/huff-kt/MainKt.class -nt src/main/kotlin/Main.kt )

huff-ada:
	( cd ../huff-ada ; alr build )
